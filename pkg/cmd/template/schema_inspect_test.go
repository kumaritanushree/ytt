// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"testing"

	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"github.com/stretchr/testify/require"
)

func TestSchemaInspect_exports_an_OpenAPI_doc(t *testing.T) {
	t.Run("for all inferred types with their inferred defaults", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
foo:
  int_key: 10
  bool_key: true
  false_key: false
  string_key: some text
  float_key: 9.1
  array_of_scalars:
  - ""
  array_of_maps:
  - foo: ""
    bar: ""
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          type: object
          additionalProperties: false
          properties:
            int_key:
              type: integer
              default: 10
            bool_key:
              type: boolean
              default: true
            false_key:
              type: boolean
              default: false
            string_key:
              type: string
              default: some text
            float_key:
              type: number
              format: float
              default: 9.1
            array_of_scalars:
              type: array
              items:
                type: string
                default: ""
              default: []
            array_of_maps:
              type: array
              items:
                type: object
                additionalProperties: false
                properties:
                  foo:
                    type: string
                    default: ""
                  bar:
                    type: string
                    default: ""
              default: []
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("including explicitly set default values", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
foo:
  #@schema/default 10
  int_key: 0
  #@schema/default True
  bool_key: false
  #@schema/default False
  false_key: true
  #@schema/default "some text"
  string_key: ""
  #@schema/default 9.1
  float_key: 0.0
  #@schema/default [1,2,3]
  array_of_scalars:
  - 0
  #@schema/default [{"bar": "thing 1"},{"bar": "thing 2"}, {"bar": "thing 3"}]
  array_of_maps:
  - bar: ""
    ree: "default"
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          type: object
          additionalProperties: false
          properties:
            int_key:
              type: integer
              default: 10
            bool_key:
              type: boolean
              default: true
            false_key:
              type: boolean
              default: false
            string_key:
              type: string
              default: some text
            float_key:
              type: number
              format: float
              default: 9.1
            array_of_scalars:
              type: array
              items:
                type: integer
                default: 0
              default:
              - 1
              - 2
              - 3
            array_of_maps:
              type: array
              items:
                type: object
                additionalProperties: false
                properties:
                  bar:
                    type: string
                    default: ""
                  ree:
                    type: string
                    default: default
              default:
              - bar: thing 1
                ree: default
              - bar: thing 2
                ree: default
              - bar: thing 3
                ree: default
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("including nullable values", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
foo:
  #@schema/nullable
  int_key: 0
  #@schema/nullable
  bool_key: false
  #@schema/nullable
  false_key: true
  #@schema/nullable
  string_key: ""
  #@schema/nullable
  float_key: 0.0
  #@schema/nullable
  array_of_scalars:
  - 0
  #@schema/nullable
  array_of_maps:
  -
    #@schema/nullable
    bar: ""
    #@schema/nullable
    ree: ""
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          type: object
          additionalProperties: false
          properties:
            int_key:
              type: integer
              nullable: true
              default: null
            bool_key:
              type: boolean
              nullable: true
              default: null
            false_key:
              type: boolean
              nullable: true
              default: null
            string_key:
              type: string
              nullable: true
              default: null
            float_key:
              type: number
              format: float
              nullable: true
              default: null
            array_of_scalars:
              type: array
              nullable: true
              items:
                type: integer
                default: 0
              default: null
            array_of_maps:
              type: array
              nullable: true
              items:
                type: object
                additionalProperties: false
                properties:
                  bar:
                    type: string
                    nullable: true
                    default: null
                  ree:
                    type: string
                    nullable: true
                    default: null
              default: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("including 'any' values", func(t *testing.T) {
		t.Run("on documents", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			opts.DataValuesFlags.InspectSchema = true
			opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

			schemaYAML := `#@data/values-schema
#@schema/type any=True
---
foo:
  int_key: 0
  array_of_scalars:
  - ""
  array_of_maps:
  - foo: ""
    bar: ""
`
			expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      nullable: true
      default:
        foo:
          int_key: 0
          array_of_scalars:
          - ""
          array_of_maps:
          - foo: ""
            bar: ""
`
			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertSucceedsDocSet(t, filesToProcess, expected, opts)
		})
		t.Run("on map items", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			opts.DataValuesFlags.InspectSchema = true
			opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

			schemaYAML := `#@data/values-schema
---
#@schema/type any=True
foo:
  int_key: 0
  array_of_scalars:
  - ""
  array_of_maps:
  - foo: ""
    bar: ""
`
			expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          nullable: true
          default:
            int_key: 0
            array_of_scalars:
            - ""
            array_of_maps:
            - foo: ""
              bar: ""
`
			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertSucceedsDocSet(t, filesToProcess, expected, opts)
		})
		t.Run("on array items", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			opts.DataValuesFlags.InspectSchema = true
			opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

			schemaYAML := `#@data/values-schema
---
foo:
  int_key: 0
  array_of_scalars:
  #@schema/type any=True
  - ""
  array_of_maps:
  - foo: ""
    bar: ""
`
			expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          type: object
          additionalProperties: false
          properties:
            int_key:
              type: integer
              default: 0
            array_of_scalars:
              type: array
              items:
                nullable: true
                default: ""
              default: []
            array_of_maps:
              type: array
              items:
                type: object
                additionalProperties: false
                properties:
                  foo:
                    type: string
                    default: ""
                  bar:
                    type: string
                    default: ""
              default: []
`
			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertSucceedsDocSet(t, filesToProcess, expected, opts)
		})
	})
	t.Run("including nullable values with defaults", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
foo:
  #@schema/default 10
  #@schema/nullable
  int_key: 0

  #@schema/default True
  #@schema/nullable
  bool_key: false

  #@schema/nullable
  #@schema/default False
  false_key: true

  #@schema/nullable
  #@schema/default "some text"
  string_key: ""

  #@schema/nullable
  #@schema/default 9.1
  float_key: 0.0

  #@schema/nullable
  #@schema/default [1,2,3]
  array_of_scalars:
  - 0

  #@schema/default [{"bar": "thing 1"},{"bar": "thing 2"}, {"bar": "thing 3"}]
  #@schema/nullable
  array_of_maps:
  -
    #@schema/nullable
    bar: ""
    #@schema/nullable
    ree: ""
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          type: object
          additionalProperties: false
          properties:
            int_key:
              type: integer
              nullable: true
              default: 10
            bool_key:
              type: boolean
              nullable: true
              default: true
            false_key:
              type: boolean
              nullable: true
              default: false
            string_key:
              type: string
              nullable: true
              default: some text
            float_key:
              type: number
              format: float
              nullable: true
              default: 9.1
            array_of_scalars:
              type: array
              nullable: true
              items:
                type: integer
                default: 0
              default:
              - 1
              - 2
              - 3
            array_of_maps:
              type: array
              nullable: true
              items:
                type: object
                additionalProperties: false
                properties:
                  bar:
                    type: string
                    nullable: true
                    default: null
                  ree:
                    type: string
                    nullable: true
                    default: null
              default:
              - bar: thing 1
                ree: null
              - bar: thing 2
                ree: null
              - bar: thing 3
                ree: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})

}
func TestSchemaInspect_annotation_adds_key(t *testing.T) {
	t.Run("in the correct relative order", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
db_conn:
  #@schema/title "Host Title"
  #@schema/desc "The hostname"
  #@schema/type any=True
  #@schema/examples ("hostname example", 1.5)
  #@schema/default "host"
  #@schema/deprecated ""
  hostname: ""
  #@schema/title "Port Title" 
  #@schema/desc "Port should be float between 0.152 through 16.35"  
  #@schema/nullable
  #@schema/examples ("", 1.5)
  #@schema/default 9.9
  #@schema/deprecated ""
  port: 0.2
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        db_conn:
          type: object
          additionalProperties: false
          properties:
            hostname:
              title: Host Title
              nullable: true
              deprecated: true
              description: The hostname
              x-example-description: hostname example
              example: 1.5
              default: host
            port:
              title: Port Title
              type: number
              format: float
              nullable: true
              deprecated: true
              description: Port should be float between 0.152 through 16.35
              x-example-description: ""
              example: 1.5
              default: 9.9
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("when description provided by @schema/desc", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
#@schema/desc "Network configuration values"
---
#@schema/desc "List of database connections"
db_conn:
#@schema/desc "A network entry"
- 
  #@schema/desc "The hostname"
  hostname: ""
  #@schema/desc "Port should be between 49152 through 65535"
  port: 0
  #@schema/desc "Timeout in minutes"
  timeout: 1.0
  #@schema/desc "Any type is allowed"
  #@schema/type any=True
  any_key: thing
  #@schema/desc "When not provided, the default is null"
  #@schema/nullable
  null_key: ""
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      description: Network configuration values
      properties:
        db_conn:
          type: array
          description: List of database connections
          items:
            type: object
            additionalProperties: false
            description: A network entry
            properties:
              hostname:
                type: string
                description: The hostname
                default: ""
              port:
                type: integer
                description: Port should be between 49152 through 65535
                default: 0
              timeout:
                type: number
                format: float
                description: Timeout in minutes
                default: 1
              any_key:
                nullable: true
                description: Any type is allowed
                default: thing
              null_key:
                type: string
                nullable: true
                description: When not provided, the default is null
                default: null
          default: []
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("when title provided by @schema/title", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
#@schema/title "Network configuration values"
---
#@schema/title "List of database connections"
db_conn:
#@schema/title "A network entry"
-
  #@schema/title "The host"
  hostname: ""
  #@schema/title "The Port"
  port: 0
  #@schema/title "The Timeout"
  timeout: 1.0
  #@schema/title "Any type"
  #@schema/type any=True
  any_key: thing
  #@schema/title "When not provided, the default is null"
  #@schema/nullable
  null_key: ""
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      title: Network configuration values
      type: object
      additionalProperties: false
      properties:
        db_conn:
          title: List of database connections
          type: array
          items:
            title: A network entry
            type: object
            additionalProperties: false
            properties:
              hostname:
                title: The host
                type: string
                default: ""
              port:
                title: The Port
                type: integer
                default: 0
              timeout:
                title: The Timeout
                type: number
                format: float
                default: 1
              any_key:
                title: Any type
                nullable: true
                default: thing
              null_key:
                title: When not provided, the default is null
                type: string
                nullable: true
                default: null
          default: []
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("when examples are provided by @schema/examples", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
#@schema/examples ("schema example description", {"db_conn": [{"hostname": "localhost", "port": 8080, "timeout": 4.2, "any_key": "anything", "null_key": None}]})
---
#@schema/examples ("db_conn example description", [{"hostname": "localhost", "port": 8080, "timeout": 4.2, "any_key": "anything", "null_key": None}])
db_conn:
#@schema/examples ("db_conn array example description", {"hostname": "localhost", "port": 8080, "timeout": 4.2, "any_key": "anything", "null_key": "not null"})
- 
  #@schema/examples ("hostname example description", "localhost")
  #@schema/desc "The hostname"
  hostname: ""
  #@schema/examples ("",8080)
  port: 0
  #@schema/examples ("timeout example description", 4.2), ("another timeout ex desc", 5)
  timeout: 1.0
  #@schema/examples ("any_key example description", "anything")
  #@schema/type any=True
  any_key: thing
  #@schema/examples ("null_key example description", None)
  #@schema/nullable
  null_key: ""
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      x-example-description: schema example description
      example:
        db_conn:
        - hostname: localhost
          port: 8080
          timeout: 4.2
          any_key: anything
          null_key: null
      properties:
        db_conn:
          type: array
          x-example-description: db_conn example description
          example:
          - hostname: localhost
            port: 8080
            timeout: 4.2
            any_key: anything
            null_key: null
          items:
            type: object
            additionalProperties: false
            x-example-description: db_conn array example description
            example:
              hostname: localhost
              port: 8080
              timeout: 4.2
              any_key: anything
              null_key: not null
            properties:
              hostname:
                type: string
                description: The hostname
                x-example-description: hostname example description
                example: localhost
                default: ""
              port:
                type: integer
                x-example-description: ""
                example: 8080
                default: 0
              timeout:
                type: number
                format: float
                x-example-description: timeout example description
                example: 4.2
                default: 1
              any_key:
                nullable: true
                x-example-description: any_key example description
                example: anything
                default: thing
              null_key:
                type: string
                nullable: true
                x-example-description: null_key example description
                example: null
                default: null
          default: []
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("when deprecated property is provided by @schema/deprecated", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
#@schema/deprecated ""
db_conn:
#@schema/deprecated ""
-
  #@schema/deprecated ""
  hostname: ""
  #@schema/deprecated ""
  port: 0
  #@schema/deprecated ""
  timeout: 1.0
  #@schema/deprecated ""
  #@schema/type any=True
  any_key: thing
  #@schema/deprecated ""
  #@schema/nullable
  null_key: ""
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        db_conn:
          type: array
          deprecated: true
          items:
            type: object
            additionalProperties: false
            deprecated: true
            properties:
              hostname:
                type: string
                deprecated: true
                default: ""
              port:
                type: integer
                deprecated: true
                default: 0
              timeout:
                type: number
                format: float
                deprecated: true
                default: 1
              any_key:
                nullable: true
                deprecated: true
                default: thing
              null_key:
                type: string
                nullable: true
                deprecated: true
                default: null
          default: []
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
}

func TestSchemaInspect_errors(t *testing.T) {
	t.Run("when --output is anything other than 'openapi-v3'", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true

		schemaYAML := `#@data/values-schema
---
foo: doesn't matter
`
		expectedErr := "Data values schema export only supported in OpenAPI v3 format; specify format with --output=openapi-v3 flag"

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})

	t.Run("when --output is set to 'openapi-v3' but not inspecting schema", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = false
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
foo: doesn't matter
`
		expectedErr := "Output type currently only supported for data values schema (i.e. include --data-values-schema-inspect)"

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
}

func assertSucceedsDocSet(t *testing.T, filesToProcess []*files.File, expectedOut string, opts *cmdtpl.Options) {
	t.Helper()
	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	require.NoError(t, out.Err)

	outBytes, err := out.DocSet.AsBytes()
	require.NoError(t, err)

	require.Equal(t, expectedOut, string(outBytes))
}
