Este proyecto es un fork de https://github.com/atlassian/gostatsd.

El propósito es añadir un nuevo "backend" a gostatsd, que envíe las métricas
a SepAgent.

Para generar el nuevo ejecutable:
- Preparar el entorno (1 única vez) con "make setup"
- Generar el ejecutable con "make build"

================================================================================
2019/04/24

Actualmente hay un problema en la preparación del entorno (make setup), durante
la instalación del paquete golangci-lint.
La ejecución de "go get -u github.com/golangci/golangci-lint/cmd/golangci-lint"
falla con un:
  go: sourcegraph.com/sourcegraph/go-diff@v0.5.1: parsing go.mod: unexpected module path "github.com/sourcegraph/go-diff"
  go get: error loading module requirements

Parece que el error está relacionado con un cambio de ubicación (sourcegraph->github)
de go-diff@v0.5.1:
https://github.com/golangci/golangci-lint/issues/497
https://github.com/golangci/golangci-lint/pull/487

Hasta que hayan aplicado la solución, podemos evitarlo así:

- Nos aseguramos de tener incluido el path a los ejecutables de go
  export PATH=$PATH:/usr/local/go/bin:/home/myname/go/bin
  (asumiendo que GOPATH = /home/myname/go)

- En el fichero gostatsd/Makefile, comentamos la línea
  # go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

- Estando fuera del directorio gostatsd (para que no encuentre el fichero go.mod),
  lanzamos la orden go get -u github.com/golangci/golangci-lint/cmd/golangci-lint.
  Nos dejará el ejecutable golangci-lint en /home/- Antes de ejecutar el "make setup", he lanzado a mano ese "go get...", FUERA del
  directorio gostatsd (si lo lanzo desde dentro, falla).
  Funciona OK, y me lo instala en /home/jvalero/go/src (no en pkg). Me da igual,
  lo que quiero es solo el ejecutable que deja en /home/myname/go/bin.

- Lanzamos el "make setup".





================================================================================