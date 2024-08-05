# 数据分发控制

## 服务端
主要文件是 `Api/Distributor.go`
服务端提供两个接口，都是get请求，**端口号及加密key可以在代码中更改，然后再编译**，

### 编译
Windows 下，linux 里一样的流程，只是安装和编译稍有不同
1. 下安装 golang1.22.0，参见 https://www.runoob.com/go/go-environment.html
2. 然后进入项目目录 `cd Distributor`
3. 执行 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 go build -o Distaribute.exe Api/Distributor.go ，即可生成 可执行文件 `distributor.exe`
4. 执行 `.\distributor.exe`
5. 浏览器访问 http://localhost:8123/get_token 有正常返回，表示ok

### 获取token接口
/get_token , 对应的c#请求代码
```c#
var client = new RestClient("http://127.0.0.1:8123/get_token");
client.Timeout = -1;
var request = new RestRequest(Method.GET);
client.UserAgent = "Apifox/1.0.0 ";
IRestResponse response = client.Execute(request);
Console.WriteLine(response.Content);
```
### 获取数据接口
/get_data, 对应的c#请求代码
```c#
var client = new RestClient("http://127.0.0.1:8123/get_data");
client.Timeout = -1;
var request = new RestRequest(Method.GET);
request.AddHeader("Authorization", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDkyMjM2NzB9.iN62a7GJQ6nVWoi5wVTfn2PUp67rV83xi7vw9--Pumg");
client.UserAgent = "Apifox/1.0.0 ";
IRestResponse response = client.Execute(request);
Console.WriteLine(response.Content);
```
### 数据解密
服务端数据加密，用的ase加密，见函数 `encryptResponse` 和 `decryptResponse` ， 对应c#解密代码
```c#
using System;
using System.Text.Json;
using System.Security.Cryptography;

// Response 是要返回的数据结构
public class Response
{
    public int Limit { get; set; }
    public string Salt { get; set; }
}

// 解密函数
public static Response DecryptResponse(byte[] cipherText, byte[] key)
{
    using (Aes aes = Aes.Create())
    {
        aes.Key = key;
        byte[] iv = new byte[aes.BlockSize / 8];
        Array.Copy(cipherText, iv, iv.Length);
        aes.IV = iv;

        using (MemoryStream memoryStream = new MemoryStream())
        {
            using (CryptoStream cryptoStream = new CryptoStream(memoryStream, aes.CreateDecryptor(), CryptoStreamMode.Write))
            {
                cryptoStream.Write(cipherText, iv.Length, cipherText.Length - iv.Length);
                cryptoStream.FlushFinalBlock();
                byte[] decryptedBytes = memoryStream.ToArray();

                // 反序列化解密后的字节流为 Response 结构
                Response response = JsonSerializer.Deserialize<Response>(decryptedBytes);
                return response;
            }
        }
    }
}
```




## 客户端
### 第一步
调用get_token接口获取token，返回是这样的
```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDkyMjM2NzB9.iN62a7GJQ6nVWoi5wVTfn2PUp67rV83xi7vw9--Pumg"}
```
### 第二步
拿这个token，调用get_data接口获取数据，返回是这样的
```json
{"data":"dmVyeXNlY3JldGtleTEyM6gIEjW65ktx6vg6uG4Q2WrXjMdxoYs2si8hKBvYROS9pOgwD9Cdvxo="}
```
这个data 要解密出来，得到其中的limit字段
解密后的数据是这样的
```json
{"limit":1000,"salt":"afdscdss23"}
```

## 签名证书制作
**生成私钥（Key）：**首先，你需要生成一个私钥文件。使用以下命令：
bash
Copy code
openssl genrsa -out server.key 2048
这会生成一个名为 server.key 的私钥文件。

**生成证书签名请求（CSR）：**接下来，你需要生成一个证书签名请求文件（CSR），用于向证书颁发机构（CA）请求签名证书。在这种情况下，我们将自己签名，所以我们将使用相同的私钥来生成 CSR。使用以下命令：
bash
Copy code
openssl req -new -key server.key -out server.csr
这会生成一个名为 server.csr 的 CSR 文件。

**生成自签名证书（CRT）：**使用私钥和 CSR 来生成自签名证书。使用以下命令：
bash
Copy code
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt
这会生成一个名为 server.crt 的自签名证书。

完成这些步骤后，你就可以将生成的证书和私钥文件用于你的 Go 服务器了

### 客户端c# 示例代码
```c#
using System;
using System.Net.Http;
using System.Net.Security;
using System.Security.Cryptography.X509Certificates;
using System.Threading.Tasks;

namespace CustomCertificateClient
{
    class Program
    {
        static async Task Main(string[] args)
        {
            // 创建 HttpClient
            HttpClientHandler handler = new HttpClientHandler();
            
            // 加载自定义签名证书
            X509Certificate2 cert = new X509Certificate2("client.pfx", "password"); // 你的客户端证书路径和密码
            handler.ClientCertificates.Add(cert);

            // 创建 HttpClient
            HttpClient client = new HttpClient(handler);

            // 设置服务器地址
            string url = "https://your_server_address/get_data"; // 替换为你的服务器地址

            try
            {
                // 发起 GET 请求
                HttpResponseMessage response = await client.GetAsync(url);

                // 读取响应
                string responseBody = await response.Content.ReadAsStringAsync();

                // 输出响应
                Console.WriteLine("Response:");
                Console.WriteLine(responseBody);
            }
            catch (Exception ex)
            {
                Console.WriteLine("Error: " + ex.Message);
            }
        }
    }
}

```

