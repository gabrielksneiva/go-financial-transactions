// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.857
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func TransactionForm() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<section><form id=\"transactionForm\" class=\"space-y-4 bg-white dark:bg-gray-800 p-6 rounded-2xl shadow-sm transition-shadow hover:shadow-md\"><div><label class=\"block mb-1 font-medium text-sm\">Tipo</label> <select name=\"type\" class=\"w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 bg-white dark:bg-gray-700 dark:text-white\" id=\"transactionType\"><option value=\"deposit\">Depósito</option> <option value=\"withdraw\">Saque</option></select></div><div><label class=\"block mb-1 font-medium text-sm\">Valor</label> <input type=\"number\" name=\"amount\" step=\"0.01\" required class=\"w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 bg-white dark:bg-gray-700 dark:text-white\"></div><button type=\"submit\" class=\"w-full bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-4 rounded-lg transition\">Enviar</button></form><script>\n    document.getElementById(\"transactionForm\")\n      .addEventListener(\"submit\", async function (ev) {\n        ev.preventDefault();\n        const formData = new FormData(ev.target);\n        const type = formData.get('type');\n        const amount = formData.get('amount');\n        const endpoint = type === \"withdraw\" ? \"/api/withdraw\" : \"/api/deposit\";\n\n        try {\n          const res = await fetch(endpoint, {\n            method: 'POST',\n            credentials: 'include',\n            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },\n            body: new URLSearchParams({ amount }),\n          });\n          if (!res.ok) throw new Error(await res.text());\n          await res.json();\n          await refreshExtract();      // <-- aqui\n        } catch (err) {\n          alert(\"Erro: \" + err.message);\n        }\n      });\n\n    async function refreshExtract() {\n      const resp = await fetch('/dashboard/extract', { credentials: 'include' });\n      if (!resp.ok) throw new Error(await resp.text());\n      const html = await resp.text();\n      document.getElementById('transactionExtract').innerHTML = html;\n      lucide.createIcons();\n    }\n\n    lucide.createIcons();\n  </script></section>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
