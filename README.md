# go_http_files
Easy http file server that allows to upload/download files.
For images server can create thumbnail on default size 300x200. 

You can send files using different ways.

### First

Use *curl* command:

```shell
curl -F "uploadFile=@/my/dir/image.jpeg" -F "pathFile=test" http://localhost:8081/upload
```

### Second

From html form:

```html
<form enctype="multipart/form-data" action="http://127.0.0.1:8081/upload" method="post">
    <input type="file" name="uploadFile" />
    <input type="text" name="pathFile" />
    <input type="submit" value="upload" />
</form>
```

### Response on request in JSON format:

```json
{
    "url":"http://127.0.0.1:8081",
    "mime":"image/jpeg",
    "original":{
        "name":"1.jpg",
        "path":"/files/test/",
        "size":273670
    },
    "thumbnail":{
        "name":"1.jpg",
        "path":"/files/test/thumbnail/",
        "size":20912
    },
    "status":"Success"}
}
```

