// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package web

var mainTpl = `
<html>

<body>

{{template "content" .}}

</body>
</html>
`

var pathTpl = `
{{define "content"}}

<form>
通道选择：
<select name="ID">
{{range $i, $v := .Content.IDMAX}}
	{{if eq 0 $i}}
		<option value={{$i}} selected="selected"}>{{$i}}</option>
	{{else}}
		<option value={{$i}}}>{{$i}}</option>
	{{end}}
{{end}}
</select>

<br/><br/>
通道名称：<input type="text" placeholder={{.Content.WorkerName}} name="PathName">
<br/><br/>
设备选择： <select name="WorkerName">
	<option value=""></option>
	{{range .Content.Workers}}
		<option value="{{.}}">{{.}}</option>
	{{end}}
</select>
<br/><br/>
<input type="submit" value="提交">
<br/><br/>
</form>

{{end}}
`
