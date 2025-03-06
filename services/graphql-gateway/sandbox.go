package main

import (
	"net/http"
)

// RenderApolloSandbox serves the Apollo Sandbox UI
func RenderApolloSandbox(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<html><body><div style="width: 100%; height: 100vh;" id='embedded-sandbox'></div>
<script src="https://embeddable-sandbox.cdn.apollographql.com/_latest/embeddable-sandbox.umd.production.min.js"></script>
<script>
  new window.EmbeddedSandbox({
    target: '#embedded-sandbox',
    initialEndpoint: 'http://localhost:8080/graphql',
  });
</script>
</body></html>
	`))
}