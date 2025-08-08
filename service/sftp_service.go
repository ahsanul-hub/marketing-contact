package service

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPService struct{}

func NewSFTPService() *SFTPService {
	return &SFTPService{}
}

type MerchantSFTPConfig struct {
	ClientName string
	AppID      string
	SFTPHost   string
	SFTPPort   string
	SFTPUser   string
	SFTPPass   string
	RemotePath string
	FileName   string
}

func (s *SFTPService) UploadFile(config MerchantSFTPConfig, fileName string, fileData []byte) error {
	log.Printf("Attempting to connect to SFTP server: %s:%s", config.SFTPHost, config.SFTPPort)

	// Buat SSH client config
	sshConfig := &ssh.ClientConfig{
		User: config.SFTPUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.SFTPPass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Koneksi ke server SFTP
	port, err := strconv.Atoi(config.SFTPPort)
	if err != nil {
		return fmt.Errorf("invalid SFTP port: %v", err)
	}

	serverAddr := fmt.Sprintf("%s:%d", config.SFTPHost, port)
	log.Printf("Connecting to SFTP server at: %s", serverAddr)

	sshClient, err := ssh.Dial("tcp", serverAddr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SFTP server: %v", err)
	}
	defer sshClient.Close()

	log.Printf("SSH connection established successfully")

	// Buat SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	log.Printf("SFTP client created successfully")

	// Buat remote file path
	remoteFilePath := fmt.Sprintf("%s%s", config.RemotePath, fileName)
	log.Printf("Uploading file to: %s", remoteFilePath)

	// Buat folder remote jika tidak ada
	err = s.createRemoteDirectory(sftpClient, config.RemotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote directory: %v", err)
	}

	// Buat file di server SFTP
	remoteFile, err := sftpClient.Create(remoteFilePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %v", err)
	}
	defer remoteFile.Close()

	// Tulis data ke file
	bytesWritten, err := remoteFile.Write(fileData)
	if err != nil {
		return fmt.Errorf("failed to write to remote file: %v", err)
	}

	log.Printf("Successfully uploaded file %s to SFTP server %s (%d bytes written)", fileName, config.SFTPHost, bytesWritten)
	return nil
}

func (s *SFTPService) createRemoteDirectory(sftpClient *sftp.Client, remotePath string) error {
	// Cek apakah folder sudah ada
	_, err := sftpClient.Stat(remotePath)
	if err == nil {
		log.Printf("Remote directory %s already exists", remotePath)
		return nil
	}

	// Jika folder tidak ada, buat folder
	log.Printf("Creating remote directory: %s", remotePath)
	err = sftpClient.MkdirAll(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote directory %s: %v", remotePath, err)
	}

	log.Printf("Successfully created remote directory: %s", remotePath)
	return nil
}

func (s *SFTPService) TestConnection(config MerchantSFTPConfig) error {
	sshConfig := &ssh.ClientConfig{
		User: config.SFTPUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.SFTPPass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	port, err := strconv.Atoi(config.SFTPPort)
	if err != nil {
		return fmt.Errorf("invalid SFTP port: %v", err)
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.SFTPHost, port), sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SFTP server: %v", err)
	}
	defer sshClient.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	// Test dengan mencoba list directory dan buat folder jika perlu
	err = s.createRemoteDirectory(sftpClient, config.RemotePath)
	if err != nil {
		return fmt.Errorf("failed to create/access remote directory: %v", err)
	}

	log.Printf("SFTP connection test successful for %s", config.SFTPHost)
	return nil
}
