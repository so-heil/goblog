// Code generated by templ - DO NOT EDIT.

// templ: version: 0.2.432
package container

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import (
	"github.com/so-heil/goblog/business/templates/components/breadcrumb"
	"github.com/so-heil/goblog/business/templates/components/contact"
	"github.com/so-heil/goblog/business/templates/components/header"
)

func traceTOC() templ.ComponentScript {
	return templ.ComponentScript{
		Name: `__templ_traceTOC_4526`,
		Function: `function __templ_traceTOC_4526(){const observer = new IntersectionObserver(entries => {
         entries.forEach(entry => {
             const id = entry.target.getAttribute('id');
             if (entry.intersectionRatio > 0) {
                 document.querySelector(` + "`" + `#toc > ul > li > a[href="#${id}"]` + "`" + `).parentElement.classList.add('active-toc-item');
             } else {
                 document.querySelector(` + "`" + `#toc > ul > li > a[href="#${id}"]` + "`" + `).parentElement.classList.remove('active-toc-item');
             }
         });
    });

    // Track all sections that have an ` + "`" + `id` + "`" + ` applied
    document.querySelectorAll('section[id]').forEach((section) => {
         observer.observe(section);
    });}`,
		Call: templ.SafeScript(`__templ_traceTOC_4526`),
	}
}

func Container(links []breadcrumb.Link) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<html><head><link rel=\"stylesheet\" href=\"/static/css/tailwind.css\"><link rel=\"stylesheet\" href=\"/static/fonts/fonts.css\"><link rel=\"stylesheet\" href=\"/static/prism/prism.css\"><script src=\"/static/prism/prism.js\"></script></head>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ.RenderScriptItems(ctx, templ_7745c5c3_Buffer, traceTOC())
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<body class=\"gradbg text-gray-300 font-mono\" onload=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 templ.ComponentScript = traceTOC()
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ_7745c5c3_Var2.Call)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><div class=\"px-6 md:px-12 flex flex-col\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = header.Header(links).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<main class=\"flex-1 min-h-[calc(100vh-105px)] flex flex-col\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</main>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = contact.Contact().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></body><style>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Var3 := `
        .gradbg {
            background: linear-gradient(-45deg, #08172E, #000000, #15253F);
            background-size: 400% 400%;
            animation: gradient 35s ease infinite;
            height: 100vh;
        }
        @keyframes gradient {
        	0% {
        		background-position: 0% 50%;
        	}
        	50% {
        		background-position: 100% 50%;
        	}
        	100% {
        		background-position: 0% 50%;
        	}
        }
        html {
        	scroll-behavior: smooth;
        }
	`
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ_7745c5c3_Var3)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</style></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
