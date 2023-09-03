#!/bin/bash

deepcopy-gen -O zz_generated.deepcopy \
  -i xcodest.me/student/pkg/apis/student/v1 \
  --go-header-file hack/boilerplate.go.txt \
  -v3 \
  --bounding-dirs . \
  --output-base ../../

register-gen -O zz_register \
  -i xcodest.me/student/pkg/apis/student/v1 \
  --go-header-file hack/boilerplate.go.txt \
  --output-base ../../
