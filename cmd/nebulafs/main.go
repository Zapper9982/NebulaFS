package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/tanmaydeobhankar/nebulafs/internal/files"
	"github.com/tanmaydeobhankar/nebulafs/internal/node"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	startPort := startCmd.Int("port", 3000, "Port to listen on")
	startPeers := startCmd.String("bootstrap", "", "Comma-separated bootstrap peers")
	startStorage := startCmd.String("storage", "./storage", "Storage directory foundation")

	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	uploadPath := uploadCmd.String("file", "", "Path to file to upload")
	uploadPort := uploadCmd.Int("port", 3001, "Port to use for temporary node")
	uploadPeers := uploadCmd.String("bootstrap", "", "Bootstrap peers")

	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	downloadMeta := downloadCmd.String("meta", "", "Path to metadata JSON file")
	downloadKey := downloadCmd.String("key", "", "Encryption key (hex)")
	downloadOut := downloadCmd.String("out", "", "Output file path")
	downloadPort := downloadCmd.Int("port", 3002, "Port to use for temporary node")
	downloadPeers := downloadCmd.String("bootstrap", "", "Bootstrap peers")

	switch os.Args[1] {
	case "start":
		startCmd.Parse(os.Args[2:])
		runNode(*startPort, *startPeers, *startStorage)
	case "upload":
		uploadCmd.Parse(os.Args[2:])
		if *uploadPath == "" {
			uploadCmd.PrintDefaults()
			os.Exit(1)
		}
		runUpload(*uploadPort, *uploadPeers, *uploadPath)
	case "download":
		downloadCmd.Parse(os.Args[2:])
		if *downloadMeta == "" || *downloadKey == "" || *downloadOut == "" {
			downloadCmd.PrintDefaults()
			os.Exit(1)
		}
		runDownload(*downloadPort, *downloadMeta, *downloadKey, *downloadOut, *downloadPeers)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: nebulafs <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  start     Start a storage node")
	fmt.Println("  upload    Upload a file")
	fmt.Println("  download  Download a file")
}

func runNode(port int, peers string, storageBase string) *node.Node {
	bootstrapList := []string{}
	if peers != "" {
		bootstrapList = strings.Split(peers, ",")
	}

	config := node.NodeConfig{
		Port:           port,
		BootstrapPeers: bootstrapList,
		StorageDir:     fmt.Sprintf("%s_%d", storageBase, port),
	}

	n, err := node.NewNode(config)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	// Only block if running as 'start' command, otherwise return node
	if len(os.Args) > 1 && os.Args[1] == "start" {
		log.Printf("Starting NebulaFS node on port %d...", port)
		if err := n.Start(); err != nil {
			log.Fatalf("Node error: %v", err)
		}
	} else {
		// Just start listeners in background
		go n.Start()
	}
	return n
}

func runUpload(port int, peers string, path string) {
	n := runNode(port, peers, "./storage")

	if peers != "" {
		fmt.Println("Waiting for bootstrap...")
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Uploading...")
	meta, key, err := n.UploadFile(path)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	// Save metadata to valid JSON to copy-paste
	metaJson, _ := json.MarshalIndent(meta, "", "  ")
	fmt.Printf("\n=== File Uploaded Successfully ===\n")
	fmt.Printf("File ID: %s\n", meta.ID)
	fmt.Printf("Key: %s\n", key)
	fmt.Printf("Metadata:\n%s\n", string(metaJson))

	// Save meta to file for convenience
	os.WriteFile(meta.Name+".meta.json", metaJson, 0644)
	fmt.Printf("Metadata saved to %s.meta.json\n", meta.Name)
}

func runDownload(port int, metaPath string, key string, out string, peers string) {
	n := runNode(port, peers, "./storage")

	if peers != "" {
		fmt.Println("Waiting for bootstrap...")
		time.Sleep(1 * time.Second)
	}

	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		log.Fatalf("Failed to read metadata: %v", err)
	}

	var meta files.FileMetadata
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		log.Fatalf("Invalid metadata: %v", err)
	}

	fmt.Println("Downloading...")
	if err := n.DownloadFile(meta, key, out); err != nil {
		log.Fatalf("Download failed: %v", err)
	}
	fmt.Printf("File downloaded to: %s\n", out)
}
