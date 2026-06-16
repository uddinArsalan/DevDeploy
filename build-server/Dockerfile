FROM node:22-slim

WORKDIR /repo-build

RUN apt-get update \ 
    && apt-get install -y git \
    && rm -rf /var/lib/apt/lists/*

COPY main.sh main.sh

COPY . .

RUN chmod +x main.sh

EXPOSE 5173 5173

ENTRYPOINT [ "./main.sh" ] 