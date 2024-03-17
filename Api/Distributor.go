package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"time"
)

// 密钥，用于签名 JWT， 线上改成自己的
var jwtKey = []byte("your_jwt_key")

// encryptKey 用于加密和解密数据， 线上改成自己的
var encryptKey = []byte("verysecretkey123")

// 定义一个自定义的声明结构体
type CustomClaims struct {
	jwt.StandardClaims
}

// Response 是要返回的数据结构
type Response struct {
	Limit int    `json:"limit"`
	Salt  string `json:"salt"`
}

func main() {
	// 定义获取token的接口
	http.HandleFunc("/get_token", getTokenHandler)
	// 定义获取data的接口
	http.HandleFunc("/get_data", getDataHandler)
	// 指定监听端口
	http.ListenAndServe(":8123", nil)
}

/**
 * @desc 获取token的具体实现
 */
func getTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 创建一个自定义声明结构体对象
	claims := CustomClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(2 * time.Minute).Unix(), // 设置过期时间为当前时间后的 2 分钟
		},
	}

	// 创建一个 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名 JWT 并获取字符串形式的令牌
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error generating token")
		return
	}

	// 返回带有 JWT 的数据
	response := map[string]interface{}{
		"token": tokenString,
	}
	json.NewEncoder(w).Encode(response)
}

/**
 * @desc 获取data的具体实现
 */
func getDataHandler(w http.ResponseWriter, r *http.Request) {
	// 从请求头中获取 JWT 令牌
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Missing authorization token")
		return
	}

	// 解析 JWT 令牌
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法是否有效
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Invalid authorization token")
		return
	}

	// 检查是否解析成功且令牌有效
	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Invalid authorization token")
		return
	}

	randomString, err := generateRandomString(16)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error generating random string:", err)
		return
	}

	// 创建要返回的数据结构
	response := Response{
		Limit: 1000,         // 这里假设要返回的数字为 1000
		Salt:  randomString, // 随机字符串，混淆加密用
	}

	// 加密整个 response
	cipherText, err := encryptResponse(response, encryptKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Encryption error:", err)
		return
	}

	// 将加密后的数据转换为 Base64 编码字符串
	encryptedData := base64.StdEncoding.EncodeToString(cipherText)

	response2 := map[string]interface{}{
		"data": encryptedData,
	}

	// 这段代码debug用，线上可去掉 解密数据
	//decodedCipherText, err := base64.StdEncoding.DecodeString(encryptedData)
	//if err != nil {
	//	fmt.Println("Decoding error:", err)
	//	return
	//}
	//decryptedResponse, err := decryptResponse(decodedCipherText, encryptKey)
	//if err != nil {
	//	fmt.Println("Decryption error")
	//	return
	//}
	//
	////输出解密后的 response
	//fmt.Println("Decrypted response:", decryptedResponse)

	json.NewEncoder(w).Encode(response2)
	return
}

// 加密函数
func encryptResponse(response Response, key []byte) ([]byte, error) {
	// 将 response 序列化为 JSON 格式的字节流
	plainText, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	// 使用 AES 加密算法对字节流进行加密
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	copy(iv, key[:aes.BlockSize])

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], plainText)
	return cipherText, nil
}

// 解密函数
func decryptResponse(cipherText []byte, key []byte) (Response, error) {
	// 使用 AES 解密算法对加密的字节流进行解密
	block, err := aes.NewCipher(key)
	if err != nil {
		return Response{}, err
	}

	if len(cipherText) < aes.BlockSize {
		return Response{}, fmt.Errorf("cipherText too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)

	// 反序列化解密后的字节流为 Response 结构
	var response Response
	if err := json.Unmarshal(cipherText, &response); err != nil {
		return Response{}, err
	}
	return response, nil
}

// 生成指定长度的随机字符串
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
