* single file upload html example
```html
<!DOCTYPE html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
		<title>upload</title>
	</head>
	<body>
		<form enctype="multipart/form-data" action="http://127.0.0.1:8001/up" method="POST">
			type:<input type="number" name="type">
			<hr/>
			group:<input type="number" name="group">
			<hr/>
			onwer:<input type="number" name="owner">
			<hr/>
			keyword:<input type="text" name="keyword">
			<hr/>
			title:<input type="text" name="title">
			<hr/>
			note:<input type="text" name="note">
			<hr/>
			<input type="file" name="file">
			<input type="hidden" name="token" value="{{.}}" />
			<hr/>
			<input type="submit" value="upload" />
		</form>
	</body>
</html>
```

---

* more files upload html example
```html
<!DOCTYPE html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
		<title>uploads</title>
	</head>
	<body>
		<form enctype="multipart/form-data" action="http://127.0.0.1:8001/ups" method="POST">
			type:<input type="number" name="type">
			<hr/>
			group:<input type="number" name="group">
			<hr/>
			owner:<input type="number" name="owner">
			<hr/>
			keyword:<input type="text" name="keyword">
			<hr/>
			title:<input type="text" name="title">
			<hr/>
			note:<input type="text" name="note">
			<hr/>
			<input type="file" name="files[]" multiple/>
			<input type="hidden" name="token" value="{{.}}" />
			<hr/>
			<input type="submit" value="upload" />
		</form>
	</body>
</html>
```
