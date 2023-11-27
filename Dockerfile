FROM golang:alpine

# Install npm
RUN apk add --update nodejs npm

# Set working directory
WORKDIR /usr/src/app

# Copy files
COPY . .

# Install dependencies
RUN npm install

# Generate css
RUN npm run css-generate

# Install template engine
RUN go install github.com/a-h/templ/cmd/templ@latest

# Generate html
RUN ${GOPATH}/bin/templ

# Expose port
EXPOSE 3000

# Run the app
CMD go run cmd/main.go && go run main.go
