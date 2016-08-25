UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
  CCFLAGS += -lOpenCL
endif
ifeq ($(UNAME_S),Darwin)
  CCFLAGS += -framework OpenCL
endif

pir: pir.c
	gcc -o $@ $^ $(CCFLAGS)
