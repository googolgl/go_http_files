# go_http_files
golang easy upload &amp; download files

Easy http file server that allows to upload/download files.

* Request:
```shell
curl -F "uploadFile=@/my/dir/image.jpeg" -F "pathFile=test" http://localhost:8081/upload
```

* Response:
```json
{
    "mime":"image/jpeg",
    "url":"http://localhost:8081//files/test/image.jpeg",
    "pathfile":"/files/test/",
    "namefile":"image.jpeg",
    "size":7285,
    "status":"Success"
}```