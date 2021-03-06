package tests

import (
	"ESFS2.0/client/common"
	"ESFS2.0/message"
	"ESFS2.0/message/protos"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"
)

func TestFileServer(t *testing.T) {

}

func TestFileClient(t *testing.T) {
	c, conn, err := common.GetFileHandleClient()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	file, _ := os.Open("auth_test.go")
	stat, _ := file.Stat()
	fileInfo := message.FileInfo{
		Name:    stat.Name(),
		Mode:    stat.Mode(),
		Size:    stat.Size(),
		ModTime: stat.ModTime(),
	}
	serializedData, _ := json.Marshal(fileInfo)

	fmt.Println(stat.Name(), stat.Mode(), stat.Size(), stat.ModTime())

	request := &protos.UploadPrepareRequest{
		Username: "memeshe",
		FileInfo: serializedData,
	}

	response, err := c.UploadPrepare(ctx, request)
	if response != nil {
		fmt.Println(response.ErrorMessage)
	}
}

func TestFileSocket(t *testing.T) {
	addr := "0.0.0.0:8959"
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	msg := message.FileSocketMessage{
		UserName: "memeshe",
		FileName: []string{"8.jpg"},
		Type:     message.FILE_UPLOAD,
	}

	serializedData, err := json.Marshal(msg)
	_, err = conn.Write(serializedData)
	if err != nil {
		fmt.Printf("socket写入数据失败 %v", err.Error())
		return
	}

	file, err := os.Open("8.jpg")
	buffer := make([]byte, 2048)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		_, err = conn.Write(buffer[:n])
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func TestListFiles(t *testing.T) {
	c, conn, err := common.GetFileHandleClient()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	msg := &protos.ListFilesRequest{
		Username: "memeshe",
	}

	response, err := c.ListFiles(ctx, msg)
	info := &message.FileInfo{}
	if response != nil {
		filesArray := response.FileInfo
		for _, data := range filesArray {
			err = json.Unmarshal(data, info)
			fmt.Println(info.Name)
		}
	}
}
