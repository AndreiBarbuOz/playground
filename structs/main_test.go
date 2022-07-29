package main

import (
	"fmt"
	"testing"
)

var (
	validGraphBytes = []byte(`{
  "graph": [
    {
      "name": "platform",
      "dependencies": [
        "objectstore",
        "cache"
      ]
    },
    {
      "name": "orchestrator",
      "dependencies": [
        "objectstore",
        "cache",
        "smbdriver"
      ]
    },
    {
      "name": "testmanager",
      "dependencies": [
        "objectstore"
      ]
    },
    {
      "name": "dataservice",
      "dependencies": [
        "objectstore"
      ]
    },
    {
      "name": "apps",
      "dependencies": [
        "objectstore"
      ]
    },
    {
      "name": "aicenter",
      "dependencies": [
        "objectstore",
        "queuestore"
      ]
    },
    {
      "name": "documentunderstanding",
      "dependencies": [
        "objectstore",
        "cache",
        "aicenter"
      ]
    },
    {
      "name": "taskmining",
      "dependencies": [
        "objectstore",
        "queuestore"
      ]
    },
    {
      "name": "processmining",
      "dependencies": [
        "objectstore",
        "queuestore"
      ]
    }
  ]
}

`)
)

var table = []struct {
	input []string
}{
	{
		input: []string{"orchestrator"},
	},
	{
		input: []string{"orchestrator", "aicenter", "documentunderstanding"},
	},
	{
		input: []string{"orchestrator", "aicenter", "documentunderstanding", "apps"},
	},
}

func BenchmarkApplicationList(b *testing.B) {
	content := validGraphBytes

	actual, _ := load(content)

	for i, v := range table {
		b.Run(fmt.Sprintf("iterations_%d", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = actual.GetApplicationList(v.input)
			}
		})
	}
}
