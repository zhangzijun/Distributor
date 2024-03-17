# 数据分发控制

## 服务端
主要文件是 `Api/Distributor.go`
服务端提供两个接口，都是get请求，**端口号及加密key可以在代码中更改，然后再编译**，

### 编译
Windows 下，linux 里一样的流程，只是安装和编译稍有不同
1. 下安装 golang1.22.0，参见 https://www.runoob.com/go/go-environment.html
2. 然后进入项目目录 `cd Distributor`
3. 执行 go.exe build -o distributor.exe .\Api\Distributor.go ，即可生成 可执行文件 `distributor.exe`
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
