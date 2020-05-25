package main

import (
	"ESFS2.0/dataserver/common"
	"ESFS2.0/message"
	"ESFS2.0/message/protos"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
)

func fileSocketServer() {
	port := 8959
	host := "0.0.0.0"
	addr := fmt.Sprintf("%s:%d", host, port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("[socket] failed to listen: %v", err.Error())
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("建立socket连接失败 %v", err.Error())
			continue
		}

		go func(conn net.Conn) {
			msg := &message.FileSocketMessage{}
			buffer := make([]byte, 2048)
			n, err := conn.Read(buffer)
			err = json.Unmarshal(buffer[:n], msg)
			if err != nil {
				log.Printf("反序列化失败 %v", err.Error())
				return
			}

			if msg.Type == message.FILE_UPLOAD {
				file, err := os.Create(path.Join(common.BaseDir, "dataserver", "data", msg.UserName, msg.FileName))
				if err != nil {
					log.Printf("打开文件失败 %v", err.Error())
					return
				}

				for {
					n, err := conn.Read(buffer)
					if err == io.EOF {
						break
					}
					_, err = file.Write(buffer[:n])
					if err != nil {
						log.Printf("写入文件失败 %v", err.Error())
						break
					}
				}
				file.Close()
			}
		}(conn)

	}

}

/**
@author js
*/
func (s *dataServer) UploadPrepare(ctx context.Context, req *protos.UploadPrepareRequest) (*protos.UploadPrepareResponse, error) {
	//反序列化，获取文件信息
	fileInfo := &message.FileInfo{}
	err := json.Unmarshal(req.FileInfo, fileInfo)
	if err != nil {
		log.Printf("反序列化失败 %v", err.Error())
		return &protos.UploadPrepareResponse{
			Ok:           false,
			ErrorMessage: protos.ErrorMessage_SERVER_ERROR,
		}, err
	}

	//创建指定文件
	//Create函数若文件已存在则会截断，不存在则新建
	file, err := os.Create(path.Join(common.BaseDir, "dataserver", "data", req.Username, fileInfo.Name))
	if err != nil {
		log.Printf("创建文件失败 %v", err.Error())
		return &protos.UploadPrepareResponse{
			Ok:           false,
			ErrorMessage: protos.ErrorMessage_SERVER_ERROR,
		}, err
	}
	defer file.Close()

	fmt.Println(fileInfo.Name, fileInfo.Size, fileInfo.Mode, fileInfo.ModTime)
	return &protos.UploadPrepareResponse{
		Ok:           true,
		ErrorMessage: protos.ErrorMessage_OK,
	}, nil
}

func (s *dataServer) ListFiles(ctx context.Context, req *protos.ListFilesRequest) (*protos.ListFilesResponse, error) {
	fileDir := path.Join(common.BaseDir, "dataserver", "data", req.Username)
	files, err := ioutil.ReadDir(fileDir)
	if err != nil {
		log.Printf("读目录失败 %v", err.Error())
		return &protos.ListFilesResponse{
			Ok:       false,
			FileInfo: nil,
		}, err
	}

	var filesArray [][]byte
	for _, f := range files {
		tmp := &message.FileInfo{
			Name:    f.Name(),
			Mode:    f.Mode(),
			Size:    f.Size(),
			ModTime: f.ModTime(),
		}
		serializedData, err := json.Marshal(tmp)
		if err != nil {
			log.Printf("反序列化文件信息失败 %v", err.Error())
			continue
		}

		filesArray = append(filesArray, serializedData)
	}

	return &protos.ListFilesResponse{
		Ok:       true,
		FileInfo: filesArray,
	}, nil
}