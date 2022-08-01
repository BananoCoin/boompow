#!/bin/bash

dlv debug --headless --listen=:2345 --api-version=2 --log app.go -- server