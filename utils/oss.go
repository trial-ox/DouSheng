package utils

import (
	"bytes"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"mime/multipart"
	"os"
)

//将视频上传到阿里云,返回视频地
func UploadVideo(fileDir string, data *multipart.FileHeader) error {
	client, err := oss.New("https://oss-cn-beijing.aliyuncs.com", "LTAI5tJ57UE7R5hW8QLeKy7X", "csSI6xPjrFiIKlHEAlREbWPQJ5yYuJ")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	// 填写存储空间名称，例如examplebucket。
	bucket, err := client.Bucket("douyin-ljy")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	file, _ := data.Open()
	err = bucket.PutObject(fileDir, file)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	return err
}

//将视频封面上传到阿里云,返回视频封面地址
func UploadPicture(fileDir string, picturebytes []byte) error {
	// 创建OSSClient实例。
	// yourEndpoint填写Bucket对应的Endpoint，以华东1（杭州）为例，填写为https://oss-cn-hangzhou.aliyuncs.com。其它Region请按实际情况填写。
	// 阿里云账号AccessKey拥有所有API的访问权限，风险很高。强烈建议您创建并使用RAM用户进行API访问或日常运维，请登录RAM控制台创建RAM用户。
	client, err := oss.New("https://oss-cn-beijing.aliyuncs.com", "LTAI5tJ57UE7R5hW8QLeKy7X", "csSI6xPjrFiIKlHEAlREbWPQJ5yYuJ")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	// 填写存储空间名称，例如examplebucket。
	bucket, err := client.Bucket("douyin-ljy")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	err = bucket.PutObject(fileDir, bytes.NewReader(picturebytes))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	return err
}
