FROM openjdk:8-slim-buster

RUN apt-get update -y

RUN apt-get install -y zlib1g zlib1g-dev curl wget apt-utils

# installations for django
RUN apt-get install -y python3 python3-dev python3-pip

# installations for node
RUN curl -sL https://deb.nodesource.com/setup_12.x | bash - && apt-get -y update

RUN apt-get install -y nodejs

# RUN apt-get install -y ruby-full

RUN apt-get install -y postgresql postgresql-contrib libpq-dev git build-essential openssl

# Installations for gradle and gradlew
RUN apt-get install gradle -y

# installations for golang
RUN curl https://storage.googleapis.com/golang/go1.14.6.linux-amd64.tar.gz -o golang.tar.gz \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz

# installations for ruby
RUN mkdir ruby-install && \
	curl -fsSL https://github.com/postmodern/ruby-install/archive/v0.6.1.tar.gz | tar --strip-components=1 -C ruby-install -xz && \
	make -C ruby-install install && \
	ruby-install --system ruby 2.7.1 && \
	gem update --system

ENV PATH /usr/local/go/bin:$PATH
