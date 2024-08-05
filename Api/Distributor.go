package Api

import (
	"Distributor/model"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"sync"
	"time"
	//_ "github.com/mattn/go-sqlite3"
	//"net/http"
	//"time"
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
	Limit     int    `json:"limit"`
	Salt      string `json:"salt"`
	Timestamp int64  `json:"timestamp"`
}

// 用于保护计数器
var counterMutex sync.Mutex
var limitCounter int = 1000

// 数据库连接
//var db *sql.DB

//func init() {
//var err error
//db, err = initDB()
//println("init")
//if err != nil {
//	fmt.Println("Error initializing database:", err)
//	return
//}
//}

func SetupServerRouter() {
	//var err error
	//db, err = initDB()
	//println("init")
	//if err != nil {
	//	fmt.Println("Error initializing database:", err)
	//	return
	//}
	println("begin server")
	//定义获取token的接口
	http.HandleFunc("/get_token", getTokenHandler)
	// 定义获取data的接口
	http.HandleFunc("/get_data", getDataHandler)
	//// 定义回掉接口，确定订单完成
	//http.HandleFunc("/call_back", getDataHandler)
	// 指定监听端口
	http.ListenAndServe(":8123", nil)
	server := &http.Server{
		Addr: ":8123",
	}
	err := server.ListenAndServeTLS("server.crt", "server.key")
	if err != nil {
		fmt.Println("ListenAndServeTLS error:", err)
	}
}

//// 初始化数据库
//func initDB() (*sql.DB, error) {
//	db, err := sql.Open("sqlite3", "./data.db")
//	if err != nil {
//		return nil, err
//	}
//	//
//	//_, err = db.Exec(`CREATE TABLE IF NOT EXISTS counters (
//	//	token TEXT PRIMARY KEY,
//	//	value INTEGER
//	//);`)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	return db, nil
//}

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

	ret, _ := model.GetData()
	if ret == false {
		fmt.Fprintln(w, "create order failed")
		return
	}

	randomString, err := generateRandomString(16)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error generating random string:", err)
		return
	}

	// 创建要返回的数据结构，包括时间戳
	response := Response{
		Limit:     limitCounter,
		Salt:      randomString, // 随机字符串，混淆加密用
		Timestamp: time.Now().Unix(),
	}

	// 加密整个 response
	//cipherText, err := encryptResponse(response, encryptKey)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	fmt.Println("Encryption error:", err)
	//	return
	//}
	model.GetData()
	// 将加密后的数据转换为 Base64 编码字符串
	// encryptedData := base64.StdEncoding.EncodeToString(cipherText)
	encryptedData := response
	response2 := map[string]interface{}{
		"data": encryptedData,
	}

	json.NewEncoder(w).Encode(response2)
	return
}

// 加密函数
//func encryptResponse(response Response, key []byte) ([]byte, error) {
//	// 将 response 序列化为 JSON 格式的字节流
//	plainText, err := json.Marshal(response)
//	if err != nil {
//		return nil, err
//	}
//
//	// 使用 AES 加密算法对字节流进行加密
//	block, err := aes.NewCipher(key)
//	if err != nil {
//		return nil, err
//	}
//
//	cipherText := make([]byte, aes.BlockSize+len(plainText))
//	iv := cipherText[:aes.BlockSize]
//	copy(iv, key[:aes.BlockSize])
//
//	cfb := cipher.NewCFBEncrypter(block, iv)
//	cfb.XORKeyStream(cipherText[aes.BlockSize:], plainText)
//	return cipherText, nil
//}

// 生成指定长度的随机字符串
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

//func getData() bool {
//	// 获取一个秘钥并创建订单
//	var secretKeyID int
//	var remainingKeys int
//
//	tx, err := db.Begin()
//	if err != nil {
//		fmt.Println("error", err.Error())
//		return false
//	}
//	defer tx.Rollback()
//
//	err = tx.QueryRow("SELECT id, remaining_keys FROM orders WHERE remaining_keys > 0 LIMIT 1").Scan(&secretKeyID, &remainingKeys)
//	if err != nil {
//		fmt.Println("error", "No available keys")
//		return false
//	}
//
//	_, err = tx.Exec("UPDATE orders SET remaining_keys = remaining_keys - 1, key_update_time = ? WHERE id = ?", time.Now(), secretKeyID)
//	if err != nil {
//		fmt.Println("error", err.Error())
//		return false
//	}
//
//	month := time.Now().Format("200601")
//	model.createOrderTableForMonth(month)
//
//	_, err = tx.Exec("INSERT INTO orders_"+month+" (task_number, mobile_platform, uid, uid_key, task_status, order_creation_time, order_update_time) VALUES (?, ?, ?, ?, ?, ?, ?)",
//		"ORDER12345", "Android", "UID123", "UIDKEY123", "burning", time.Now(), time.Now())
//	if err != nil {
//		fmt.Println("error", err.Error())
//		return false
//	}
//
//	err = tx.Commit()
//	if err != nil {
//		fmt.Println("error", err.Error())
//		return false
//	}
//
//	return true
//}

//func callBack(c *gin.Context) {
//	var orderStatus struct {
//		OrderNumber string `json:"task_number" binding:"required"`
//		Status      string `json:"status" binding:"required"`
//		Reason      string `json:"reason"`
//	}
//
//	if err := c.ShouldBindJSON(&orderStatus); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//
//	month := time.Now().Format("200601")
//	tableName := "orders_" + month
//
//	tx, err := db.Begin()
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	defer tx.Rollback()
//
//	if orderStatus.Status == "burning successful" {
//		_, err = tx.Exec("UPDATE orders SET successful_keys = successful_keys + 1, key_update_time = ? WHERE id = (SELECT secret_key_id FROM "+tableName+" WHERE task_number = ?)", time.Now(), orderStatus.OrderNumber)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//	}
//
//	_, err = tx.Exec("UPDATE "+tableName+" SET task_status = ?, order_update_time = ?, remarks = ? WHERE task_number = ?",
//		orderStatus.Status, time.Now(), orderStatus.Reason, orderStatus.OrderNumber)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	err = tx.Commit()
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"message": "Order updated"})
//}
