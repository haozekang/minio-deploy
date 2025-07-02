# appsettings.json
```json
{
  "Minio": {
    "Endpoint": "192.168.120.142:9000",
    "Region": "cn-north-1",
    "AccessKey": "superadmin",
    "SecretKey": "superadmin",
    "TempUploadBucket": "temp-upload"
  }
}
```

# Program.cs
```cs
public class Program
{
    public static string MinioEndpoint = string.Empty;
    public static string MinioRegion = string.Empty;
    public static string MinioAccessKey = string.Empty;
    public static string MinioSecretKey = string.Empty;
    public static string TempUploadBucket = string.Empty;

    public static void Main(string[] args)
    {
        var builder = WebApplication.CreateBuilder(args);
        MinioEndpoint = builder.Configuration["Minio:Endpoint"] ?? string.Empty;
        MinioRegion = builder.Configuration["Minio:Region"] ?? string.Empty;
        MinioAccessKey = builder.Configuration["Minio:AccessKey"] ?? string.Empty;
        MinioSecretKey = builder.Configuration["Minio:SecretKey"] ?? string.Empty;
        TempUploadBucket = builder.Configuration["Minio:TempUploadBucket"] ?? string.Empty;

        builder.Services.AddMinio(configureClient =>
            configureClient
                .WithEndpoint(MinioEndpoint)
                .WithRegion(MinioRegion)
                .WithCredentials(MinioAccessKey, MinioSecretKey)
                .WithSSL(false)
                .Build()
        );
	}
}
```

# FileController

```cs
/// <summary>
/// 测试文件上传功能，根据etag判断文件是否成功上传，文件上传成功返回guid数据
/// </summary>
public class FileController : Controller
{
	private readonly IMinioClient minioClient;
	
	public FileController(IMinioClient _minioClient) 
	{
		minioClient = _minioClient;
	}
	
	[Route("/Api/Model/v1/Test/File/Upload")]
	[HttpPost]
	[RequestFormLimits(ValueLengthLimit = 100 * 1024 * 1024, MultipartBodyLengthLimit = 100 * 1024 * 1024, ValueCountLimit = 10)]
	public async Task<IActionResult> TestUpload()
	{
		ContentResult result = new ContentResult();
		dynamic res = new ExpandoObject();
		try
		{
			if (HttpContext.Request.ContentLength == 0)
			{
				throw new TvException(-1, "服务端没有收到任何有效数据！");
			}
			if (!HttpContext.Request.HasFormContentType)
			{
				throw new TvException(50002, "Error Content-Type!");
			}
			var files = HttpContext.Request?.Form?.Files;
			if (files == null || files.Count == 0)
			{
				throw new TvException(50004, "表单无文件数据！");
			}
			var file = files.FirstOrDefault(f => f.Name == "upload_file");
			if (file == null || file.Length == 0)
			{
				throw new TvException(50001, "文件读取异常！");
			}
			var beArgs = new BucketExistsArgs().WithBucket(Program.TempUploadBucket);
			bool found = minioClient.BucketExistsAsync(beArgs).Result;
			if (!found)
			{
				throw new TvException(
					50004,
					$"未发现有效的Bucket【{Program.TempUploadBucket}】！"
				);
			}
			var id = Guid.NewGuid();
			var putObjectArgs = new PutObjectArgs()
					.WithBucket(Program.TempUploadBucket)
					.WithObject($"{id}")
					.WithStreamData(file.OpenReadStream())
					.WithObjectSize(file.Length);
			var resp = await minioClient
				.PutObjectAsync(putObjectArgs).ConfigureAwait(false);
			if (resp.Etag == null)
			{
				throw new TvException(
					50001,
					$"系统异常！"
				);
			}
			res.data = id;
			result.StatusCode = (int)HttpStatusCode.OK;
		}
		catch (MinioException ex)
		{
			result.StatusCode = (int)HttpStatusCode.Forbidden;
			res.code = 1;
			res.msg = ex.Message;
		}
		catch (Exception ex)
		{
			result.StatusCode = (int)HttpStatusCode.Forbidden;
			Debug.WriteLine(ex);
			res.code = 50001;
			res.msg = ex.Message;
		}
		result.ContentType = "application/json;charset=UTF-8";
		result.Content = JsonSerializer.Serialize(res);
		return result;
	}
}
```