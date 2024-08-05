package Api

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func test() {
	// 加载自定义签名证书
	cert, err := ioutil.ReadFile("client.crt")
	if err != nil {
		fmt.Println("Error reading certificate:", err)
		return
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cert)

	// 创建 TLS 配置
	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: true, // 忽略服务端证书验证，因为是自签名证书
	}

	// 创建自定义的 Transport，使用 TLS 配置
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// 创建客户端
	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second, // 设置超时时间
	}

	// 发起 GET 请求
	resp, err := client.Get("https://your_server_address/get_data")
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Println("Response:", string(body))
}
