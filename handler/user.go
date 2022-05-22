package handler

import (
	"context"
	"crypto/sha512"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"shop_srvs/user_srv/global"
	"shop_srvs/user_srv/model"
	"shop_srvs/user_srv/proto"
	"strings"
	"time"
)

type UserServer struct{}

// Paginate 分页实现
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// ModelToResponse User模型转化为Proto Resp
func ModelToResponse(user model.User) *proto.UserInfoResponse {

	userInfoRsp := &proto.UserInfoResponse{
		Id:       user.ID,
		Password: user.Password,
		Mobile:   user.Mobile,
		NickName: user.NickName,
		Gender:   user.Gender,
		Role:     int32(user.Role),
	}
	// 由于birthday字段注册时没有要求填写
	// grpc中字段赋值不能是nil
	// 所以进行额外检查
	if user.Birthday != nil {
		userInfoRsp.Birthday = uint64(user.Birthday.Unix())
	}

	return userInfoRsp
}

// GetUserList 获取用户列表
func (s *UserServer) GetUserList(ctx context.Context, req *proto.PageInfo) (*proto.UserListResponse, error) {

	var (
		users []model.User
		res   *gorm.DB
	)

	if res = global.DB.Find(&users); res.Error != nil {
		return nil, res.Error
	}

	rsp := &proto.UserListResponse{
		Total: int32(res.RowsAffected),
	}

	global.DB.Scopes(Paginate(int(req.Pn), int(req.PSize))).Find(&users)

	for _, user := range users {
		userInfoRsp := ModelToResponse(user)
		rsp.Data = append(rsp.Data, userInfoRsp)
	}
	return rsp, nil
}

// GetUserByMobile 通过手机号码查询用户
func (s *UserServer) GetUserByMobile(ctx context.Context, req *proto.MobileRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	res := global.DB.Where(&model.User{Mobile: req.Mobile}).First(&user)
	if res.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	if res.Error != nil {
		return nil, status.Error(codes.Internal, res.Error.Error())
	}
	return ModelToResponse(user), nil
}

// GetUserById 根据用户ID查询用户
func (s *UserServer) GetUserById(ctx context.Context, req *proto.IdRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	res := global.DB.First(&user, req.Id)
	if res.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	if res.Error != nil {
		return nil, status.Error(codes.Internal, res.Error.Error())
	}
	return ModelToResponse(user), nil
}

// CreateUser 创建用户
func (s *UserServer) CreateUser(ctx context.Context, req *proto.CreateUserInfo) (*proto.UserInfoResponse, error) {
	var user model.User
	// 创建前查看用户是否存在
	res := global.DB.Where(&model.User{Mobile: req.Mobile}).First(&user)
	if res.RowsAffected > 0 {
		return nil, status.Errorf(codes.AlreadyExists, "用户已存在")
	}

	// 密码加密
	// Using custom options
	options := &password.Options{SaltLen: 10, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}
	salt, encodedPwd := password.Encode(req.PassWord, options)

	// 构造用户数据
	user = model.User{
		BaseModel: model.BaseModel{},
		Mobile:    req.Mobile,
		Password:  fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd),
		NickName:  req.NickName,
	}

	res = global.DB.Create(&user)
	if res.Error != nil {
		return nil, status.Errorf(codes.Internal, res.Error.Error())
	}
	return ModelToResponse(user), nil
}

// UpdateUser 更新用户
func (s *UserServer) UpdateUser(ctx context.Context, req *proto.UpdateUserInfo) (*empty.Empty, error) {
	var user model.User
	res := global.DB.First(&user, req.Id)
	if res.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	// 构造用户数据
	birthDay := time.Unix(int64(req.Birthday), 0)

	user = model.User{
		NickName: req.NickName,
		Birthday: &birthDay,
		Gender:   req.Gender,
	}
	res = global.DB.Save(user)
	if res.Error != nil {
		return nil, status.Error(codes.Internal, res.Error.Error())
	}
	return &empty.Empty{}, nil
}

// CheckPassWord 校验密码准确性
func (s *UserServer) CheckPassWord(ctx context.Context, req *proto.PasswordCheckInfo) (*proto.CheckResponse, error) {
	// 校验密码
	options := &password.Options{SaltLen: 10, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}
	// 切割密码获取salt和Encrypted password
	passwordInfo := strings.Split(req.EncryptedPassword, "$")

	// 根据原始密码和salt+EncryptedPassword验证密码有效性
	ok := password.Verify(req.Password, passwordInfo[2], passwordInfo[3], options)
	return &proto.CheckResponse{Success: ok}, nil
}
