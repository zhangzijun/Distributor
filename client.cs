using System;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Threading.Tasks;
using System.Text.Json;

public class Program
{
    private static readonly HttpClient client = new HttpClient();

    public static async Task Main(string[] args)
    {
        string token = await GetTokenAsync();
        if (string.IsNullOrEmpty(token))
        {
            Console.WriteLine("Failed to retrieve token");
            return;
        }

        string data = await GetDataAsync(token);
        if (string.IsNullOrEmpty(data))
        {
            Console.WriteLine("Failed to retrieve data");
            return;
        }

        // 解密数据
        string decryptedData = DecryptData(data);

        // 解析 JSON
        DataResponse response = JsonSerializer.Deserialize<DataResponse>(decryptedData);

        // 输出 limit 字段
        Console.WriteLine($"Limit: {response.Limit}");
    }

    public static async Task<string> GetTokenAsync()
    {
        HttpResponseMessage response = await client.GetAsync("http://localhost:8123/get_token");
        if (response.IsSuccessStatusCode)
        {
            string tokenJson = await response.Content.ReadAsStringAsync();
            TokenResponse tokenResponse = JsonSerializer.Deserialize<TokenResponse>(tokenJson);
            return tokenResponse.Token;
        }
        else
        {
            return null;
        }
    }

    public static async Task<string> GetDataAsync(string token)
    {
        client.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", token);
        HttpResponseMessage response = await client.GetAsync("http://localhost:8123/get_data");
        if (response.IsSuccessStatusCode)
        {
            return await response.Content.ReadAsStringAsync();
        }
        else
        {
            return null;
        }
    }

    public static string DecryptData(string data)
    {
        // 这里写解密逻辑，示例中假设直接返回解密后的字符串
        // 在实际应用中，需要使用相同的加密算法和密钥进行解密
        // 并处理异常情况
        return data;
    }
}

// 定义 TokenResponse 类，用于解析获取 token 的返回数据
public class TokenResponse
{
    public string Token { get; set; }
}

// 定义 DataResponse 类，用于解析获取数据的返回数据
public class DataResponse
{
    public int Limit { get; set; }
}
