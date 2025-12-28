// Package handlers contains HTTP handlers for the API.
package handlers

import (
	"html/template"
	"net/http"
)

const playgroundTemplate = `<!DOCTYPE html>
<html>
<head>
  <title>AI Aggregator - GraphQL Playground</title>
  <link rel="stylesheet" href="https://unpkg.com/graphiql/graphiql.min.css" />
  <style>
    body {
      height: 100%;
      margin: 0;
      width: 100%;
      overflow: hidden;
    }
    #graphiql {
      height: 100vh;
    }
  </style>
</head>
<body>
  <div id="graphiql">Loading...</div>
  <script crossorigin src="https://unpkg.com/react/umd/react.production.min.js"></script>
  <script crossorigin src="https://unpkg.com/react-dom/umd/react-dom.production.min.js"></script>
  <script crossorigin src="https://unpkg.com/graphiql/graphiql.min.js"></script>
  <script>
    const fetcher = GraphiQL.createFetcher({
      url: '{{.Endpoint}}',
      headers: {
        'X-API-Key': 'dev-api-key-12345'
      }
    });

    ReactDOM.render(
      React.createElement(GraphiQL, { fetcher: fetcher }),
      document.getElementById('graphiql'),
    );
  </script>
</body>
</html>`

// PlaygroundHandler returns the GraphQL playground.
func PlaygroundHandler(endpoint string) http.HandlerFunc {
	tmpl := template.Must(template.New("playground").Parse(playgroundTemplate))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, map[string]string{
			"Endpoint": endpoint,
		})
	}
}
