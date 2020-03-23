// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package web

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/zhanglongx/Aqua/comm"
	"github.com/zhanglongx/Aqua/manager"
)

// M is shortcut for map
type M map[string]interface{}

var ep = &manager.EPath

func init() {

	if err := ep.Create("testdata/test1.json"); err != nil {
		comm.Error.Panicf("Create EncodePath failed")
	}

	http.HandleFunc("/", pathIdx)

	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func pathIdx(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	IDStr := r.Form.Get("ID")

	data := make(map[interface{}]interface{})

	content := pathGet(IDStr)

	content["IDMAX"] = []int{0, 1, 2, 3}
	content["Workers"] = ep.GetWorkers()

	if IDStr == "" {
		content["ID"] = 0
	} else {
		content["ID"], _ = strconv.Atoi(IDStr)
	}

	data["Content"] = content

	execTpl(w, data, pathTpl)
}

func pathGet(IDStr string) M {

	if IDStr == "" {
		return make(M)
	}

	id, _ := strconv.Atoi(IDStr)

	var params manager.Params
	var err error
	if params, err = ep.Get(id); err != nil {
		comm.Error.Panicf("Get path %d failed", id)
		return make(M)
	}

	return M(params)
}

// beego: https://github.com/astaxie/beego
func execTpl(rw http.ResponseWriter, data map[interface{}]interface{}, tpls ...string) {
	tmpl := template.Must(template.New("main").Parse(mainTpl))
	for _, tpl := range tpls {
		tmpl = template.Must(tmpl.Parse(tpl))
	}
	tmpl.Execute(rw, data)
}
