#!/bin/bash

curl -X POST \
  'https://identitytoolkit.googleapis.com/v1/accounts:signUp?key=AIzaSyBLxg7i0kDP9yYoGM2LhQwhQg6mAB-uVI0' \
  -H 'Content-Type: application/json' \
  -d '{"returnSecureToken": true}'
