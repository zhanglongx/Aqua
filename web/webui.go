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

var epTpl = `
{{define "content"}}

<form>

* 首次启动后数据为fake，需要至少先查询一次参数
<br></br>
* 每次设置参数前都需要先查询参数
<br></br>

通道选择：
<select name="ID">
	{{range $k, $v := .Content.ID}}
		{{if eq $k 0}}
			<option value={{$v}} selected="selected"}>{{$v}}</option>
		{{else}}
			<option value={{$v}}>{{$v}}</option>
		{{end}}
	{{end}}
</select>
<br></br>

通道名称：
{{with $pn := .Content.PathName}}
	<input type="text" name="PathName" value={{$pn}}>
{{else}}
	<input type="text" name="PathName">
{{end}}
<br></br>

设备选择： 
<select name="WorkerName">
	{{range $k, $v := .Content.WorkerName}}
		{{if eq $k 0}}
			<option value={{$v}} selected="selected"}>{{$v}}</option>
		{{else}}
			<option value={{$v}}>{{$v}}</option>
		{{end}}
	{{end}}
</select>
<br></br>

RTSP地址：
<input type="text" name="rtsp_url" value={{.Content.Card.rtsp_url}}>
<br></br>

是否启动：
{{if eq .Content.IsRunning true}}
	<input type="checkbox" name="IsRunning" value=1 checked>
{{else}}
	<input type="checkbox" name="IsRunning" value=1>
{{end}}
<br></br>

<input type="submit" name="get" value="查询参数">
<input type="submit" name="set" value="设置参数">
<br></br>

</form>

{{range $e := .Content.Error}} {{$e}}<br></br> {{end}}

{{end}}
`

var dpTpl = `
{{define "content"}}

<form>

* 首次启动后数据为fake，需要至少先查询一次参数
<br></br>
* 每次设置参数前都需要先查询参数
<br></br>

通道选择：
<select name="ID">
	{{range $k, $v := .Content.ID}}
		{{if eq $k 0}}
			<option value={{$v}} selected="selected"}>{{$v}}</option>
		{{else}}
			<option value={{$v}}>{{$v}}</option>
		{{end}}
	{{end}}
</select>
<br></br>

设备选择： 
<select name="WorkerName">
	{{range $k, $v := .Content.WorkerName}}
		{{if eq $k 0}}
			<option value={{$v}} selected="selected"}>{{$v}}</option>
		{{else}}
			<option value={{$v}}>{{$v}}</option>
		{{end}}
	{{end}}
</select>
<br></br>

是否启动：
{{if eq .Content.IsRunning true}}
	<input type="checkbox" name="IsRunning" value=1 checked>
{{else}}
	<input type="checkbox" name="IsRunning" value=1>
{{end}}
<br></br>

<input type="submit" name="get" value="查询参数">
<input type="submit" name="set" value="设置参数">
<br></br>

</form>

{{range $e := .Content.Error}} {{$e}}<br></br> {{end}}

{{end}}
`
