#!/bin/sh
echo Broccoli Bot
curl -X POST -F "images_file=@food.jpg" \
                "https://gateway-a.watsonplatform.net/visual-recognition/api/v3/classify?api_key={831c7f1589067b00e90a98bbd423d47780d6ddd0}&version=2016-05-20" -o ./assets/response.json