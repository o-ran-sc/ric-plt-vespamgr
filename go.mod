module gerrit.o-ran-sc.org/r/ric-plt/vespamgr

go 1.13

replace gerrit.o-ran-sc.org/r/ric-plt/xapp-frame => gerrit.o-ran-sc.org/r/ric-plt/xapp-frame.git v0.4.0

replace gerrit.o-ran-sc.org/r/ric-plt/sdlgo => gerrit.o-ran-sc.org/r/ric-plt/sdlgo.git v0.5.2

replace gerrit.o-ran-sc.org/r/com/golog => gerrit.o-ran-sc.org/r/com/golog.git v0.0.2

require (
	gerrit.o-ran-sc.org/r/com/golog.git v0.0.1
	gerrit.o-ran-sc.org/r/ric-plt/xapp-frame v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.4.0
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
)
