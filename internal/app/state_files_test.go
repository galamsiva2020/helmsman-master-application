package app

import (
	"os"
	"reflect"
	"testing"
)

func setupStateFileTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")
	os.MkdirAll(tempFilesDir, 0o755)

	return func(t *testing.T) {
		t.Log("teardown test case")
		os.RemoveAll(tempFilesDir)
	}
}

func Test_fromTOML(t *testing.T) {
	type args struct {
		file string
		s    *state
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test case 1 -- invalid TOML",
			args: args{
				file: "../../../tests/invalid_example.toml",
				s:    new(state),
			},
			want: false,
		}, {
			name: "test case 2 -- valid TOML",
			args: args{
				file: "../../examples/example.toml",
				s:    new(state),
			},
			want: true,
		},
	}
	os.Setenv("ORG_PATH", "sample")
	os.Setenv("VALUE", "sample")

	teardownTestCase := setupStateFileTestCase(t)
	defer teardownTestCase(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.s.fromTOML(tt.args.file)
			got := err == nil
			if got != tt.want {
				t.Errorf("fromToml() = %v, want %v", got, tt.want)
			}
		})
	}
	os.Unsetenv("ORG_PATH")
	os.Unsetenv("VALUE")
}

func Test_fromTOML_Expand(t *testing.T) {
	type args struct {
		file string
		s    *state
	}
	tests := []struct {
		name    string
		args    args
		section string
		field   string
		want    string
	}{
		{
			name: "test case 1 -- valid TOML expand ClusterURI",
			args: args{
				file: "../../examples/example.toml",
				s:    new(state),
			},
			section: "Settings",
			field:   "ClusterURI",
			want:    "https://192.168.99.100:8443",
		},
		{
			name: "test case 2 -- valid TOML expand org",
			args: args{
				file: "../../examples/example.toml",
				s:    new(state),
			},
			section: "Metadata",
			field:   "org",
			want:    "example.com/sample/",
		},
	}
	os.Setenv("SET_URI", "https://192.168.99.100:8443")
	os.Setenv("ORG_PATH", "sample")
	os.Setenv("VALUE", "sample")

	teardownTestCase := setupStateFileTestCase(t)
	defer teardownTestCase(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.s.fromTOML(tt.args.file)
			if err != nil {
				t.Errorf("fromToml(), got: %v", err)
			}

			tomlVal := reflect.ValueOf(tt.args.s).Elem()
			tomlType := reflect.TypeOf(tt.args.s)

			if tomlType.Kind() != reflect.Struct {

				section := tomlVal.FieldByName(tt.section)
				sectionType := reflect.TypeOf(section)

				if section.IsValid() && section.Kind() == reflect.Struct {
					field := section.FieldByName(tt.field)
					if sectionType.Kind() == reflect.String {
						if field.String() != tt.want {
							t.Errorf("fromToml().section.field = %v, got: %v", tt.want, field.String())
						}
					}
				} else if section.IsValid() && section.Kind() == reflect.Map {
					found := false
					value := ""
					for _, key := range section.MapKeys() {
						if key.String() == tt.field {
							found = true
							value = section.MapIndex(key).String()
						}
					}
					if !found {
						t.Errorf("fromToml().section.field = '%v' not found", tt.field)
					} else if value != tt.want {
						t.Errorf("fromToml().section.field = %v, got: %v", tt.want, value)
					}

				} else {
					t.Errorf("fromToml().section = struct, got: %v", sectionType.Kind())
				}

			} else {
				t.Errorf("fromToml() = struct, got: %v", tomlType.Kind())
			}
		})
	}
	os.Unsetenv("ORG_PATH")
	os.Unsetenv("SET_URI")
	os.Unsetenv("VALUE")
}

func Test_fromYAML(t *testing.T) {
	type args struct {
		file string
		s    *state
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test case 1 -- invalid YAML",
			args: args{
				file: "../../tests/invalid_example.yaml",
				s:    new(state),
			},
			want: false,
		}, {
			name: "test case 2 -- valid TOML",
			args: args{
				file: "../../examples/example.yaml",
				s:    new(state),
			},
			want: true,
		},
	}
	os.Setenv("VALUE", "sample")
	os.Setenv("ORG_PATH", "sample")

	teardownTestCase := setupStateFileTestCase(t)
	defer teardownTestCase(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.s.fromYAML(tt.args.file)
			got := err == nil
			if got != tt.want {
				t.Errorf("fromYaml() = %v, want %v", got, tt.want)
			}
		})
	}
	os.Unsetenv("ORG_PATH")
	os.Unsetenv("VALUE")
}

func Test_fromYAML_UnsetVars(t *testing.T) {
	type args struct {
		file string
		s    *state
	}
	tests := []struct {
		name      string
		args      args
		targetVar string
		want      bool
	}{
		{
			name: "test case 1 -- unset ORG_PATH env var",
			args: args{
				file: "../../examples/example.yaml",
				s:    new(state),
			},
			targetVar: "ORG_PATH",
			want:      false,
		},
		{
			name: "test case 2 -- unset VALUE var",
			args: args{
				file: "../../examples/example.yaml",
				s:    new(state),
			},
			targetVar: "VALUE",
			want:      false,
		},
	}

	teardownTestCase := setupStateFileTestCase(t)
	defer teardownTestCase(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.targetVar == "ORG_PATH" {
				os.Setenv("VALUE", "sample")
			} else if tt.targetVar == "VALUE" {
				os.Setenv("ORG_PATH", "sample")
			}
			err := tt.args.s.fromYAML(tt.args.file)
			got := err == nil
			if got != tt.want {
				t.Errorf("fromYaml() = %v, want %v", got, tt.want)
			}
		})
		os.Unsetenv("ORG_PATH")
		os.Unsetenv("VALUE")
	}
}

func Test_fromYAML_Expand(t *testing.T) {
	type args struct {
		file string
		s    *state
	}
	tests := []struct {
		name    string
		args    args
		section string
		field   string
		want    string
	}{
		{
			name: "test case 1 -- valid YAML expand ClusterURI",
			args: args{
				file: "../../examples/example.yaml",
				s:    new(state),
			},
			section: "Settings",
			field:   "ClusterURI",
			want:    "https://192.168.99.100:8443",
		},
		{
			name: "test case 2 -- valid YAML expand org",
			args: args{
				file: "../../examples/example.yaml",
				s:    new(state),
			},
			section: "Metadata",
			field:   "org",
			want:    "example.com/sample/",
		},
	}
	os.Setenv("SET_URI", "https://192.168.99.100:8443")
	os.Setenv("ORG_PATH", "sample")
	os.Setenv("VALUE", "sample")

	teardownTestCase := setupStateFileTestCase(t)
	defer teardownTestCase(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.s.fromYAML(tt.args.file)
			if err != nil {
				t.Errorf("fromYaml(), got: %v", err)
			}

			yamlVal := reflect.ValueOf(tt.args.s).Elem()
			yamlType := reflect.TypeOf(tt.args.s)

			if yamlType.Kind() != reflect.Struct {

				section := yamlVal.FieldByName(tt.section)
				sectionType := reflect.TypeOf(section)

				if section.IsValid() && section.Kind() == reflect.Struct {
					field := section.FieldByName(tt.field)
					if sectionType.Kind() == reflect.String {
						if field.String() != tt.want {
							t.Errorf("fromYaml().section.field = %v, got: %v", tt.want, field.String())
						}
					}
				} else if section.IsValid() && section.Kind() == reflect.Map {
					found := false
					value := ""
					for _, key := range section.MapKeys() {
						if key.String() == tt.field {
							found = true
							value = section.MapIndex(key).String()
						}
					}
					if !found {
						t.Errorf("fromYaml().section.field = '%v' not found", tt.field)
					} else if value != tt.want {
						t.Errorf("fromYaml().section.field = %v, got: %v", tt.want, value)
					}

				} else {
					t.Errorf("fromYaml().section = struct, got: %v", sectionType.Kind())
				}

			} else {
				t.Errorf("fromYaml() = struct, got: %v", yamlType.Kind())
			}
		})
	}
	os.Unsetenv("ORG_PATH")
	os.Unsetenv("SET_URI")
	os.Unsetenv("VALUE")
}
