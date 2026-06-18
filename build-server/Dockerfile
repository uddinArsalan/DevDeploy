# From base docker image add git 
FROM dhi.io/docker:29-cli-dev

WORKDIR /repo-build

RUN apt-get update \ 
    && apt-get install -y git curl tar gzip \
    && rm -rf /var/lib/apt/lists/*

# copy main.sh and other project dependencies
COPY main.sh main.sh

COPY . .

# chmod to executable (when we do docker run entrypoint will be executed first in this case
# main.sh content which clone repo and cd into directory and then prepare railpack and build image)
RUN chmod +x main.sh

# RUN docker pull ghcr.io/railwayapp/railpack-frontend:v0.27.2
RUN curl -sSL https://railpack.com/install.sh | sh -s -- --yes

ENTRYPOINT [ "./main.sh" ] 