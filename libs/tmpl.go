package libs

import "text/template"

var GptTmplStr = `
---
services:
  pg:
    image: {{ .ImagePG }}
    container_name: pg
    restart: always
#    ports:
#      - 5432:5432
    networks:
      - fastgpt
    environment:
      - POSTGRES_USER={{ .DbUser }}
      - POSTGRES_PASSWORD={{ .DbPass }}
      - POSTGRES_DB=postgres
    volumes:
      - {{ .DataDir }}/pg:/var/lib/postgresql/data
      - {{ .BackupDir }}/pg:/backup

  mongo:
    image:  {{ .ImageMG }}
    container_name: mongo
    restart: always
    ports:
      - 27017:27017
    networks:
      - fastgpt
    command: mongod --keyFile /data/mongodb.key --replSet rs0
    environment:
      - MONGO_INITDB_ROOT_USERNAME={{ .DbUser }}
      - MONGO_INITDB_ROOT_PASSWORD={{ .DbPass }}
    volumes:
      - {{ .DataDir }}/mongo:/data/db
      - {{ .BackupDir }}/mongo:/backup
    entrypoint:
      - bash
      - -c
      - |
        openssl rand -base64 128 > /data/mongodb.key
        chmod 400 /data/mongodb.key
        chown 999:999 /data/mongodb.key
        echo 'const isInited = rs.status().ok === 1
        if(!isInited){
          rs.initiate({
              _id: "rs0",
              members: [
                  { _id: 0, host: "mongo:27017" }
              ]
          })
        }' > /data/initReplicaSet.js
        exec docker-entrypoint.sh "$$@" &
        until mongo -u {{ .DbUser }} -p {{ .DbPass }} --authenticationDatabase admin --eval "print('waited for connection')" > /dev/null 2>&1; do
          echo "Waiting for MongoDB to start..."
          sleep 2
        done
        mongo -u {{ .DbUser }} -p {{ .DbPass }} --authenticationDatabase admin /data/initReplicaSet.js
        wait $$!

  fastgpt:
    container_name: fastgpt
    image: {{ .ImageGPT }}
    ports:
      - 3100:3000
    networks:
      - fastgpt
    depends_on:
      - mongo
      - pg
      - oneapi
      - sandbox
    restart: always
    environment:
      - DEFAULT_ROOT_PSW={{ .GptPass }}
      - OPENAI_BASE_URL={{ .BaseURL }}
      - CHAT_API_KEY={{ .ApiKey }}
      - ROOT_KEY={{ .RootKey }}
      - DB_MAX_LINK=30
      - TOKEN_KEY=any
      - FILE_TOKEN_KEY=filetoken
      - MONGODB_URI=mongodb://{{ .DbUser }}:{{ .DbPass }}@mongo:27017/fastgpt?authSource=admin
      - PG_URL=postgresql://{{ .DbUser }}:{{ .DbPass }}@pg:5432/postgres
      - LOG_LEVEL=info
      - SANDBOX_URL=http://sandbox:3000
      - STORE_LOG_LEVEL=warn
    volumes:
      - {{ .DataDir }}/config.json:/app/data/config.json

  sandbox:
    container_name: sandbox
    image: {{ .ImageSD }}
    networks:
      - fastgpt
    restart: always

  mysql:
    image: {{ .ImageMySql }}
    container_name: mysql
    restart: always
    ports:
      - 3306:3306
    networks:
      - fastgpt
    command: --default-authentication-plugin=mysql_native_password
    environment:
      MYSQL_ROOT_PASSWORD: {{ .DbPass }}
      MYSQL_DATABASE: oneapi
    volumes:
      - {{ .DataDir }}/mysql:/var/lib/mysql
      - {{ .BackupDir }}/mysql:/backup

  oneapi:
    image: {{ .ImageAPI }}
    container_name: oneapi
    restart: always
    ports:
      - "3200:3000"
    depends_on:
      - mysql
    networks:
      - fastgpt
    environment:
      - SQL_DSN=root:{{ .DbPass }}@tcp(mysql:3306)/oneapi
      - INITIAL_ROOT_TOKEN={{ .ApiKey }}
      - INITIAL_ROOT_ACCESS_TOKEN={{ .RootKey }}
      - SESSION_SECRET=oneapikey
      - MEMORY_CACHE_ENABLED=true
      - BATCH_UPDATE_ENABLED=true
      - BATCH_UPDATE_INTERVAL=10
      - TZ=Asia/Shanghai
    volumes:
      - {{ .DataDir }}/oneapi:/data

  ngx:
    image: 'nginx:alpine'
    container_name: nginx
    restart: unless-stopped
    ports:
      - '80:80'
      - '443:443'
    depends_on:
      - oneapi
      - fastgpt
    networks:
      - fastgpt
    volumes:
      - {{ .DataDir }}/nginx/nginx.conf:/etc/nginx/nginx.conf
      - {{ .DataDir }}/nginx/conf.d:/etc/nginx/conf.d
      - {{ .DataDir }}/nginx/www:/var/www
      - {{ .DataDir }}/nginx/log:/var/log/nginx

networks:
  fastgpt:
`

func GetTemp() map[string]*template.Template {
	Temps := make(map[string]*template.Template)
	Temps["gpt"], _ = template.New("gpt").Parse(GptTmplStr)
	return Temps
}

var NgxConfig = `
include /etc/nginx/module.d/*.module;

user nobody nogroup;
worker_processes  auto;

#pid        logs/nginx.pid;
#error_log  logs/error.log;
#error_log  logs/error.log  notice;
#error_log  logs/error.log  info;

events {
    worker_connections  8192;
}

include /etc/nginx/conf.d/*.conf;
`

var GptConfig = `
{
  "feConfigs": {
    "lafEnv": "https://laf.dev"
  },
  "systemEnv": {
    "vectorMaxProcess": 15,
    "qaMaxProcess": 15,
    "pgHNSWEfSearch": 100
  },
  "llmModels": [
    {
      "model": "gpt-4o-mini",
      "name": "gpt-4o-mini",
      "avatar": "/imgs/model/openai.svg",
      "maxContext": 125000,
      "maxResponse": 4000,
      "quoteMaxToken": 120000,
      "maxTemperature": 1.2,
      "charsPointsPrice": 0,
      "censor": false,
      "vision": true,
      "datasetProcess": true,
      "usedInClassify": true,
      "usedInExtractFields": true,
      "usedInToolCall": true,
      "usedInQueryExtension": true,
      "toolChoice": true,
      "functionCall": false,
      "customCQPrompt": "",
      "customExtractPrompt": "",
      "defaultSystemChatPrompt": "",
      "defaultConfig": {}
    },
    {
      "model": "gpt-4o",
      "name": "gpt-4o",
      "avatar": "/imgs/model/openai.svg",
      "maxContext": 125000,
      "maxResponse": 4000,
      "quoteMaxToken": 120000,
      "maxTemperature": 1.2,
      "charsPointsPrice": 0,
      "censor": false,
      "vision": true,
      "datasetProcess": false,
      "usedInClassify": true,
      "usedInExtractFields": true,
      "usedInToolCall": true,
      "usedInQueryExtension": true,
      "toolChoice": true,
      "functionCall": false,
      "customCQPrompt": "",
      "customExtractPrompt": "",
      "defaultSystemChatPrompt": "",
      "defaultConfig": {}
    },
    {
      "model": "yi-large",
      "name": "yi-large",
      "avatar": "/imgs/model/yi.svg",
      "maxContext": 32000,
      "maxResponse": 4000,
      "quoteMaxToken": 30000,
      "maxTemperature": 1.2,
      "charsPointsPrice": 0,
      "censor": false,
      "vision": true,
      "datasetProcess": false,
      "usedInClassify": true,
      "usedInExtractFields": true,
      "usedInToolCall": true,
      "usedInQueryExtension": true,
      "toolChoice": true,
      "functionCall": false,
      "customCQPrompt": "",
      "customExtractPrompt": "",
      "defaultSystemChatPrompt": "",
      "defaultConfig": {}
    },
   {
      "model": "yi-vision",
      "name": "yi-vision",
      "avatar": "/imgs/model/yi.svg",
      "maxContext": 16000,
      "maxResponse": 4000,
      "quoteMaxToken": 16000,
      "maxTemperature": 1,
      "charsPointsPrice": 0,
      "censor": false,
      "vision": true,
      "datasetProcess": false,
      "usedInClassify": false,
      "usedInExtractFields": false,
      "usedInToolCall": false,
      "usedInQueryExtension": false,
      "toolChoice": false,
      "functionCall": false,
      "customCQPrompt": "",
      "customExtractPrompt": "",
      "defaultSystemChatPrompt": "",
      "defaultConfig": {}
    },
    {
      "model": "glm-4-air",
      "name": "glm-4-air",
      "avatar": "/imgs/model/chatglm.svg",
      "maxContext": 125000,
      "maxResponse": 4000,
      "quoteMaxToken": 120000,
      "maxTemperature": 1.2,
      "charsPointsPrice": 0,
      "censor": false,
      "vision": true,
      "datasetProcess": true,
      "usedInClassify": true,
      "usedInExtractFields": true,
      "usedInToolCall": true,
      "usedInQueryExtension": true,
      "toolChoice": true,
      "functionCall": false,
      "customCQPrompt": "",
      "customExtractPrompt": "",
      "defaultSystemChatPrompt": "",
      "defaultConfig": {}
    },
    {
      "model": "glm-4",
      "name": "glm-4",
      "avatar": "/imgs/model/glm.svg",
      "maxContext": 125000,
      "maxResponse": 4000,
      "quoteMaxToken": 120000,
      "maxTemperature": 1.2,
      "charsPointsPrice": 0,
      "censor": false,
      "vision": true,
      "datasetProcess": false,
      "usedInClassify": true,
      "usedInExtractFields": true,
      "usedInToolCall": true,
      "usedInQueryExtension": true,
      "toolChoice": true,
      "functionCall": false,
      "customCQPrompt": "",
      "customExtractPrompt": "",
      "defaultSystemChatPrompt": "",
      "defaultConfig": {}
    },
    {
      "model": "glm-4v",
      "name": "glm-4v",
      "avatar": "/imgs/model/chatglm.svg",
      "maxContext": 32000,
      "maxResponse": 4000,
      "quoteMaxToken": 30000,
      "maxTemperature": 1.2,
      "charsPointsPrice": 0,
      "censor": false,
      "vision": true,
      "datasetProcess": false,
      "usedInClassify": true,
      "usedInExtractFields": true,
      "usedInToolCall": true,
      "usedInQueryExtension": true,
      "toolChoice": true,
      "functionCall": false,
      "customCQPrompt": "",
      "customExtractPrompt": "",
      "defaultSystemChatPrompt": "",
      "defaultConfig": {}
    },
    {
      "model": "step-1v-32k",
      "name": "step-1v-32k",
      "avatar": "/imgs/model/stepchat.svg",
      "maxContext": 32000,
      "maxResponse": 4000,
      "quoteMaxToken": 32000,
      "maxTemperature": 1.2,
      "charsPointsPrice": 0,
      "censor": false,
      "vision": true,
      "datasetProcess": false,
      "usedInClassify": true,
      "usedInExtractFields": true,
      "usedInToolCall": true,
      "usedInQueryExtension": true,
      "toolChoice": true,
      "functionCall": false,
      "customCQPrompt": "",
      "customExtractPrompt": "",
      "defaultSystemChatPrompt": "",
      "defaultConfig": {}
    },
    {
      "model": "step-1-128k",
      "name": "step-1-128k",
      "avatar": "/imgs/model/stepchat.svg",
      "maxContext": 125000,
      "maxResponse": 4000,
      "quoteMaxToken": 120000,
      "maxTemperature": 1.2,
      "charsPointsPrice": 0,
      "censor": false,
      "vision": false,
      "datasetProcess": false,
      "usedInClassify": true,
      "usedInExtractFields": true,
      "usedInToolCall": true,
      "usedInQueryExtension": true,
      "toolChoice": true,
      "functionCall": false,
      "customCQPrompt": "",
      "customExtractPrompt": "",
      "defaultSystemChatPrompt": "",
      "defaultConfig": {}
    },
    {
      "model": "moonshot-v1-128k",
      "name": "moonshot-v1-128k",
      "avatar": "/imgs/model/moonshot.svg",
      "maxContext": 125000,
      "maxResponse": 8000,
      "quoteMaxToken": 120000,
      "maxTemperature": 1.2,
      "charsPointsPrice": 0,
      "censor": false,
      "vision": true,
      "datasetProcess": true,
      "usedInClassify": true,
      "usedInExtractFields": true,
      "usedInToolCall": true,
      "usedInQueryExtension": true,
      "toolChoice": true,
      "functionCall": false,
      "customCQPrompt": "",
      "customExtractPrompt": "",
      "defaultSystemChatPrompt": "",
      "defaultConfig": {}
    }
  ],
  "vectorModels": [
    {
      "model": "text-embedding-v1",
      "name": "Embedding-Qwen-V1",
      "avatar": "/imgs/model/qwen.svg",
      "charsPointsPrice": 0,
      "defaultToken": 700,
      "maxToken": 3000,
      "weight": 100,
      "defaultConfig": {},
      "dbConfig": {},
      "queryConfig": {}
    },
    {
      "model": "embedding-2",
      "name": "Zhipu-Embedding-2",
      "avatar": "/imgs/model/chatglm.svg",
      "charsPointsPrice": 0,
      "defaultToken": 512,
      "maxToken": 3000,
      "weight": 100,
      "defaultConfig": {
	    "dimension":1024
	  },
      "dbConfig": {},
      "queryConfig": {}
    },
    {
      "model": "embo-01",
      "name": "Embedding-Minimax",
      "charsPointsPrice": 0,
      "defaultToken": 1024,
      "maxToken": 4096,
      "weight": 100,
      "dbConfig": {},
      "queryConfig": {}
        },
    {
      "model": "text-embedding-3-large",
      "name": "text-embedding-3-large",
      "avatar": "/imgs/model/openai.svg",
      "charsPointsPrice": 0,
      "defaultToken": 512,
      "maxToken": 3000,
      "weight": 100,
      "defaultConfig": {
        "dimensions": 1024
      }
    },
    {
      "model": "text-embedding-async-v1",
      "name": "text-embedding-async-v1",
      "avatar": "/imgs/model/qwen.svg",
      "charsPointsPrice": 0,
      "defaultToken": 512,
      "maxToken": 3000,
      "weight": 100
    }
  ],
  "reRankModels": [    	
   {
            "model": "jina-reranker-v2-base-multilingual",
            "name": "jina", 
            "requestUrl": "https://api.jina.ai/v1/rerank",
            "requestAuth": "jina_2d3f208e8029405e8661abb9adcfddc8XpINy7NtD91azAhrZEcSb35jHSJz",
            "top-n": 6
    }
  ],
  "audioSpeechModels": [
    {
      "model": "tts-1",
      "name": "OpenAI TTS1",
      "charsPointsPrice": 0,
      "voices": [
        { "label": "Alloy", "value": "alloy", "bufferId": "openai-Alloy" },
        { "label": "Echo", "value": "echo", "bufferId": "openai-Echo" },
        { "label": "Fable", "value": "fable", "bufferId": "openai-Fable" },
        { "label": "Onyx", "value": "onyx", "bufferId": "openai-Onyx" },
        { "label": "Nova", "value": "nova", "bufferId": "openai-Nova" },
        { "label": "Shimmer", "value": "shimmer", "bufferId": "openai-Shimmer" }
      ]
    }
  ],
  "whisperModel": {
    "model": "whisper-1",
    "name": "Whisper1",
    "charsPointsPrice": 0
  }
}
`
