* single file upload html example
```html
<html>
	<head>
		<title>upload</title>
	</head>
	<body>
		<form enctype="multipart/form-data" action="http://127.0.0.1:8001/up" method="POST">
			<input type="file" name="file">
			<input type="hidden" name="token" value="{{.}}" />
			<input type="submit" value="upload" />
		</form>
	</body>
</html>
```

---

* more files upload html example
```html
<html>
	<head>
		<title>uploads</title>
	</head>
	<body>
		<form enctype="multipart/form-data" action="http://127.0.0.1:8001/ups" method="POST">
			<input type="file" name="files[]" multiple/>
			<input type="hidden" name="token" value="{{.}}" />
			<input type="submit" value="upload" />
		</form>
	</body>
</html>
```
