# go_http_files
Easy http file server that allows to upload/download files.
You can send files using different ways.

##First
Use *curl* command

* Request:
```shell
curl -F "uploadFile=@/my/dir/image.jpeg" -F "pathFile=test" http://localhost:8081/upload
```

* Response:
```json
{
    "mime":"image/jpeg",
    "url":"http://localhost:8081/files/test/image.jpeg",
    "pathfile":"/files/test/",
    "namefile":"image.jpeg",
    "size":7285,
    "status":"Success"
}

```

##Second
From html form

```html
<form enctype="multipart/form-data" action="http://127.0.0.1:8081/upload" method="post">
    <input type="file" name="uploadFile" />
    <input type="text" name="pathFile" />
    <input type="submit" value="upload" />
</form>
```