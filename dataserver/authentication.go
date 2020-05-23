package main

import (
	"ESFS2.0/dataserver/common"
	"ESFS2.0/protos"
	"ESFS2.0/utils"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path"
)

/**
@author js
*/
func (s *dataServer) Login(ctx context.Context, req *protos.LoginRequest) (*protos.LoginResponse, error) {
	//获取数据库连接
	db, err := common.GetDBConnection()
	if err != nil {
		log.Printf("连接数据库失败 %v", err.Error())
		return &protos.LoginResponse{
			Ok:           false,
			ErrorMessage: protos.AuthErrorMessage_SERVER_ERROR}, err
	}

	//查询用户是否存在
	sql := fmt.Sprintf("select password_hash,salt from users where username='%s'", req.Username)
	res, err := common.DoQuery(sql, db)
	if err != nil {
		log.Printf("查询数据库失败 %v", err.Error())
		return &protos.LoginResponse{
			Ok:           false,
			ErrorMessage: protos.AuthErrorMessage_SERVER_ERROR}, err
	}

	var passwordHash, salt string
	exists := false
	if res.Next() {
		exists = true
		res.Scan(&passwordHash, salt)
	}
	if exists == false {
		return &protos.LoginResponse{
			Ok:           false,
			ErrorMessage: protos.AuthErrorMessage_USER_NOT_EXISTS}, nil
	}

	//检查密码是否正确
	user_passwordHash := utils.HashWithSalt(req.Password, salt)
	if user_passwordHash == passwordHash {
		return &protos.LoginResponse{
			Ok:           true,
			ErrorMessage: protos.AuthErrorMessage_OK,
		}, nil
	}

	return &protos.LoginResponse{
		Ok:           false,
		ErrorMessage: protos.AuthErrorMessage_PASSWORD_WRONG,
	}, nil
}

/**
@author js
*/
func (s *dataServer) Register(ctx context.Context, req *protos.RegisterRequest) (*protos.RegisterResponse, error) {
	//获取数据库连接
	db, err := common.GetDBConnection()
	if err != nil {
		log.Printf("连接数据库失败 %v", err.Error())
		return &protos.RegisterResponse{
			Ok:           false,
			ErrorMessage: protos.AuthErrorMessage_SERVER_ERROR}, err
	}

	//查询用户是否存在
	sql := fmt.Sprintf("select username from users where username='%s'", req.Username)
	res, err := common.DoQuery(sql, db)
	if err != nil {
		log.Printf("查询数据库失败 %v", err.Error())
		return &protos.RegisterResponse{
			Ok:           false,
			ErrorMessage: protos.AuthErrorMessage_SERVER_ERROR}, err
	}

	if res.Next() { //如果已存在，则返回失败
		return &protos.RegisterResponse{
			Ok:           false,
			ErrorMessage: protos.AuthErrorMessage_USER_ALREADY_EXISTS,
		}, nil
	}

	username := req.Username
	password := req.Password
	defaultSecondKey := req.DefaultSecondKey
	salt := base64.URLEncoding.EncodeToString(utils.GenerateRandomBytes(32)) //随机生成salt
	passwordHash := utils.HashWithSalt(password, salt)
	sql = fmt.Sprintf("insert into users values('%s','%s','%s','%s')", username, passwordHash, salt, defaultSecondKey)

	//向数据库中插入数据
	_, err = common.DoExecTx(sql, db)
	if err != nil {
		log.Printf("数据库执行插入事务失败 %v", err.Error())
		return &protos.RegisterResponse{
			Ok:           false,
			ErrorMessage: protos.AuthErrorMessage_SERVER_ERROR,
		}, nil
	}

	//创建用户文件目录
	err = os.Mkdir(path.Join(common.BaseDir, "dataserver", "data", username), os.ModePerm)
	if err != nil {
		log.Printf("创建目录失败 %v", err.Error())
		return &protos.RegisterResponse{
			Ok:           false,
			ErrorMessage: protos.AuthErrorMessage_SERVER_ERROR,
		}, err
	}

	return &protos.RegisterResponse{
		Ok:           true,
		ErrorMessage: protos.AuthErrorMessage_OK,
	}, nil
}
