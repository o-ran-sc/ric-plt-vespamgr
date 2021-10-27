module gerrit.o-ran-sc.org/r/ric-plt/vespamgr

go 1.16

replace gerrit.o-ran-sc.org/r/ric-plt/xapp-frame => gerrit.o-ran-sc.org/r/ric-plt/xapp-frame.git v0.9.1

replace gerrit.o-ran-sc.org/r/ric-plt/sdlgo => gerrit.o-ran-sc.org/r/ric-plt/sdlgo.git v0.7.0

replace gerrit.o-ran-sc.org/r/com/golog => gerrit.o-ran-sc.org/r/com/golog.git v0.0.2

require (
    gerrit.o-ran-sc.org/r/ric-plt/xapp-frame v0.9.1
    github.com/stretchr/testify v1.5.1
    gopkg.in/yaml.v2 v2.3.0
)
