package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/sftp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/ssh"
)

type LogEntry struct {
	Date            string `bson:"date" json:"date"`
	Time            string `bson:"time" json:"time"`
	ExecutionServer string `bson:"execution_server" json:"execution_server"`
	TargetServer    string `bson:"target_server" json:"target_server"`
}

var (
	servers        map[string]string
	sshKeyPath     string
	username       string
	remoteDir      string
	localDir       string
	mongoURI       string
	dbName         string
	collectionName string
	dbCollection   *mongo.Collection
	logPattern     = regexp.MustCompile(`(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2}) successfull connection from (\w+)`)
)

func init() {
	serversJSON := getEnv("SERVERS", `{"sftp1":"192.168.33.11","sftp2":"192.168.33.12","sftp3":"192.168.33.13"}`)
	if err := json.Unmarshal([]byte(serversJSON), &servers); err != nil {
		log.Fatalf("Failed to parse SERVERS: %v", err)
	}

	sshKeyPath = getEnv("SSH_KEY_PATH", ".\\my_sftp_key")
	username = getEnv("USERNAME", "sftpuser")
	remoteDir = getEnv("REMOTE_DIR", "/home/sftpuser/")
	localDir = getEnv("LOCAL_DIR", "downloaded_logs")
	mongoURI = getEnv("MONGO_URI", "mongodb://root:example@localhost:27017")
	dbName = getEnv("DB_NAME", "logdb")
	collectionName = getEnv("COLLECTION_NAME", "logs")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	log.Println("Starting application...")
	ctx := context.Background()
	os.MkdirAll(localDir, os.ModePerm)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB")
	dbCollection = client.Database(dbName).Collection(collectionName)

	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/collect", collectHandler)
	r.HandleFunc("/download", downloadCSVHandler)
	r.HandleFunc("/graph", graphHandler)
	r.HandleFunc("/graph-data", graphDataHandler)

	log.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling index request")
	ctx := context.Background()

	opts := options.Find().SetSort(bson.D{
		{Key: "date", Value: -1},
		{Key: "time", Value: -1},
	}).SetLimit(250)

	cursor, err := dbCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		log.Printf("Failed to fetch logs: %v", err)
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	var results []LogEntry
	if err := cursor.All(ctx, &results); err != nil {
		log.Printf("Failed to decode logs: %v", err)
		http.Error(w, "Failed to decode logs", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, results)
}

func collectHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling collect request")
	for name, ip := range servers {
		log.Printf("Processing server: %s (%s)", name, ip)
		if err := downloadAndProcessLogs(name, ip); err != nil {
			log.Printf("Error with %s: %v", name, err)
		}
	}
	fmt.Fprint(w, "<p>Logs collected successfully. <a href='/'>Go back</a></p>")
}

func downloadAndProcessLogs(name, ip string) error {
	log.Printf("Connecting to SFTP server: %s (%s)", name, ip)
	key, err := os.ReadFile(sshKeyPath)
	if err != nil {
		log.Printf("Failed to read SSH key: %v", err)
		return err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Printf("Failed to parse SSH key: %v", err)
		return err
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	conn, err := ssh.Dial("tcp", ip+":22", config)
	if err != nil {
		log.Printf("Failed to connect to SFTP server: %v", err)
		return err
	}
	defer conn.Close()

	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		log.Printf("Failed to create SFTP client: %v", err)
		return err
	}
	defer sftpClient.Close()

	files, err := sftpClient.ReadDir(remoteDir)
	if err != nil {
		log.Printf("Failed to read remote directory: %v", err)
		return err
	}

	errChan := make(chan error, len(files))
	var wg sync.WaitGroup

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".log") {
			wg.Add(1)
			go func(fileName string) {
				defer wg.Done()
				log.Printf("Processing log file: %s", fileName)
				remotePath := remoteDir + fileName
				localPath := filepath.Join(localDir, name+"_"+fileName)

				src, err := sftpClient.Open(remotePath)
				if err != nil {
					log.Printf("Failed to open remote file (%s): %v", remotePath, err)
					errChan <- err
					return
				}
				defer src.Close()

				data, err := io.ReadAll(src)
				if err != nil {
					log.Printf("Failed to read remote file: %v", err)
					errChan <- err
					return
				}
				if err := os.WriteFile(localPath, data, 0644); err != nil {
					log.Printf("Failed to write local file: %v", err)
					errChan <- err
					return
				}
				processLogFile(localPath, name)
			}(f.Name())
		}
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			log.Printf("Error during processing: %v", err)
		}
	}

	return nil
}

func processLogFile(path, targetServer string) {
	log.Printf("Processing log file: %s", path)
	ctx := context.Background()
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Failed to read log file: %v", err)
		return
	}
	lines := strings.Split(string(data), "\n")

	existingEntries := make(map[string]struct{})
	cursor, err := dbCollection.Find(ctx, bson.M{"target_server": targetServer})
	if err != nil {
		log.Printf("Failed to fetch existing entries: %v", err)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var entry LogEntry
		if err := cursor.Decode(&entry); err != nil {
			log.Printf("Failed to decode entry: %v", err)
			continue
		}
		key := fmt.Sprintf("%s_%s_%s_%s", entry.Date, entry.Time, entry.ExecutionServer, entry.TargetServer)
		existingEntries[key] = struct{}{}
	}

	var bulkEntries []interface{}
	for _, line := range lines {
		matches := logPattern.FindStringSubmatch(line)
		if len(matches) == 4 {
			entry := LogEntry{
				Date:            matches[1],
				Time:            matches[2],
				ExecutionServer: matches[3],
				TargetServer:    targetServer,
			}
			key := fmt.Sprintf("%s_%s_%s_%s", entry.Date, entry.Time, entry.ExecutionServer, entry.TargetServer)
			if _, exists := existingEntries[key]; exists {
				continue
			}
			bulkEntries = append(bulkEntries, entry)
			existingEntries[key] = struct{}{}
		}
	}

	if len(bulkEntries) > 0 {
		_, err := dbCollection.InsertMany(ctx, bulkEntries)
		if err != nil {
			log.Printf("Failed to insert log entries in bulk: %v", err)
		} else {
			log.Printf("Successfully inserted %d log entries in bulk", len(bulkEntries))
		}
	}
}

func downloadCSVHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling CSV download request")
	ctx := context.Background()
	cursor, err := dbCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to fetch logs: %v", err)
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	var results []LogEntry
	if err := cursor.All(ctx, &results); err != nil {
		log.Printf("Failed to decode logs: %v", err)
		http.Error(w, "Failed to decode logs", http.StatusInternalServerError)
		return
	}

	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)
	writer.Write([]string{"date", "time", "execution_server", "target_server"})
	for _, row := range results {
		writer.Write([]string{row.Date, row.Time, row.ExecutionServer, row.TargetServer})
	}
	writer.Flush()

	w.Header().Set("Content-Disposition", "attachment;filename=logs.csv")
	w.Header().Set("Content-Type", "text/csv")
	w.Write(buf.Bytes())
}

func graphHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling graph request")
	tmpl, err := template.ParseFiles("templates/graph.html")
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func graphDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling graph data request")
	ctx := context.Background()
	cursor, err := dbCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to fetch logs: %v", err)
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	var results []LogEntry
	if err := cursor.All(ctx, &results); err != nil {
		log.Printf("Failed to decode logs: %v", err)
		http.Error(w, "Failed to decode logs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
