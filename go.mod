module github.com/osrg/gobgp/v3

require (
	github.com/BurntSushi/toml v1.0.0
	github.com/Microsoft/go-winio v0.5.2
	github.com/coreos/go-systemd/v22 v22.3.2
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13
	github.com/eapache/channels v1.1.0
	github.com/go-test/deep v1.0.8
	github.com/google/go-cmp v0.5.8
	github.com/google/uuid v1.3.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/k-sone/critbitgo v1.4.0
	github.com/kr/pretty v0.3.0
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.8.0
	github.com/vishvananda/netlink v1.1.1-0.20210330154013-f5de75959ad5
	golang.org/x/net v0.0.0-20220909164309-bea034e7d591
	golang.org/x/text v0.3.7
	google.golang.org/grpc v1.49.0
	google.golang.org/protobuf v1.28.1
	infiot.com/infiot/dplane/gogwutils v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.6.1 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae // indirect
	golang.org/x/sys v0.0.0-20220728004956-3c1f35247d10 // indirect
	google.golang.org/genproto v0.0.0-20220923205249-dd2d53f1fffc // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

//replace github.com/osrg/gobgp => ../gobgp
//replace github.com/osrg/gobgp/internal/pkg/simplefpm => ../gobgp/internal/pkg/simplefpm
replace infiot.com/infiot/dplane/gogwutils => ../dplane/gogwutils

go 1.18
