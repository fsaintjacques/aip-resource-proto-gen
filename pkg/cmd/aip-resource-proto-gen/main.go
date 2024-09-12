package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stoewer/go-strcase"
)

type Config struct {
	Syntax         string
	Resource       string
	PluralResource string
	Package        string
	Service        string
	Methods        string

	// Flags controlling the generated resource

	// Whether the resource id is required in the Create/Update methods
	IDRequired bool
	// Pattern of the parent resource, if any
	ParentPattern string
	// Whether to generate the display_name field for resource name
	WithDisplayName bool
	// Whether to generate fields for resource name and create/update timestamps
	WithTimestamps bool
	// Whether to generate the annotations field
	WithAnnotations bool

	// Flags controlling the generated methods

	// Whether to generate the order_by field for list method
	WithListOrderBy bool
	// Whether to generate the filter field for list method
	WithListFilter bool
	// Whether to generate the update_mask field for update method
	WithUpdateFieldMask bool
	// Whether to generate the allow_missing field for update method
	WithUpdateAllowMissing bool
	// Whether to generate the allow_missing field for delete method
	WithDeleteAllowMissing bool

	// Flags controlling the generated options

	// Whether to generate HTTP-specific options to methods and service
	WithHTTPOptions bool

	Compact bool
}

func (c *Config) HasParent() bool {
	return c.ParentPattern != ""
}

func (c *Config) ResourceCollectionIdentifier() string {
	return strcase.LowerCamelCase(c.PluralResource)
}

func (c *Config) ResourceNamePattern() string {
	parts := []string{
		c.ParentPattern,
		c.ResourceCollectionIdentifier(),
		fmt.Sprintf("{%s}", c.ResourceSnakeCase()),
	}

	if !c.HasParent() {
		parts = parts[1:]
	}

	return strings.Join(parts, "/")
}

var replaceCurly = regexp.MustCompile(`\{([^}]+)\}`)

func (c *Config) ResourceNameUrlRef() string {
	return replaceCurly.ReplaceAllString(c.ResourceNamePattern(), "*")
}

func (c *Config) ResourceTypeName() string {
	return fmt.Sprintf("%s/%s", c.Service, c.Resource)
}

func (c *Config) ResourceSnakeCase() string {
	return strcase.SnakeCase(c.Resource)
}

func (c *Config) PluralResourceSnakeCase() string {
	return strcase.SnakeCase(c.PluralResource)
}

func (c *Config) ParentNameUrlRef() string {
	return replaceCurly.ReplaceAllString(c.ParentPattern, "*")
}

func main() {
	var cfg Config

	var cmd = &cobra.Command{
		Use:   "aip-resource-proto-gen <resource>",
		Short: "Scaffold protobuf IDL file for AIP resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Resource = args[0]

			if cfg.PluralResource == "" {
				cfg.PluralResource = cfg.Resource + "s"
			}

			s := &schemaBuilder{cfg: &cfg}

			desc, err := s.Build()
			if err != nil {
				return fmt.Errorf("failed to generate file descriptor: %v", err)
			}

			printer := initPrinter(&cfg)

			return printer.PrintProtoFile(desc, os.Stdout)

		},
	}

	// Resource flags
	cmd.Flags().StringVar(&cfg.PluralResource, "resource-plural", "", "Plural form of the resource name")
	cmd.Flags().StringVar(&cfg.ParentPattern, "resource-parent", "", "Pattern of the parent resource, if any")
	cmd.Flags().BoolVar(&cfg.IDRequired, "resource-id-required", false, "Whether the resource id is required in the Create/Update methods")
	cmd.Flags().BoolVar(&cfg.WithDisplayName, "resource-with-display-name", true, "Whether to generate the display_name field for resource")
	cmd.Flags().BoolVar(&cfg.WithTimestamps, "resource-with-timestamps", true, "Whether to generate fields for resource name and create/update timestamps")
	cmd.Flags().BoolVar(&cfg.WithAnnotations, "resource-with-annotations", true, "Whether to generate the annotations field for the resource")

	cmd.Flags().StringVar(&cfg.Package, "package", "", "Package name for the generated protobuf file")
	cmd.MarkFlagRequired("package")
	cmd.Flags().StringVar(&cfg.Service, "service", "", "Service name for the generated protobuf file")
	cmd.MarkFlagRequired("service")

	cmd.Flags().StringVar(&cfg.Syntax, "syntax", "proto3", "Syntax for the generated protobuf file")

	cmd.Flags().StringVar(&cfg.Methods, "methods", "crudl", "Comma-separated list of methods to generate")

	cmd.Flags().BoolVar(&cfg.WithHTTPOptions, "with-http-options", true, "Generate HTTP-specific options")
	cmd.Flags().BoolVar(&cfg.WithListOrderBy, "with-list-order-by", true, "Generate the order_by field for list method")
	cmd.Flags().BoolVar(&cfg.WithListFilter, "with-list-filter", true, "Generate the filter field for list method")
	cmd.Flags().BoolVar(&cfg.WithUpdateFieldMask, "with-update-field-mask", true, "Generate the update_mask field for update method")
	cmd.Flags().BoolVar(&cfg.WithUpdateAllowMissing, "with-update-allow-missing", true, "Generate the allow_missing field for update method")
	cmd.Flags().BoolVar(&cfg.WithDeleteAllowMissing, "with-delete-allow-missing", true, "Generate the allow_missing field for delete method")

	cmd.Flags().BoolVar(&cfg.Compact, "compact", false, "Generate compact proto file")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
