FROM golang:alpine

# Install npm
RUN apk add --update nodejs npm

# Set working directory
WORKDIR /usr/src/app

# Copy the Go modules and sum files first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of code
COPY . .

# Install template engine
RUN go install github.com/a-h/templ/cmd/templ@latest

# Generate html
RUN ${GOPATH}/bin/templ generate

# Install dependencies
RUN npm install

# Generate css
RUN npm run css-generate

RUN go build -o /usr/src/app/main .

# Expose port
EXPOSE 3000

# Run the app
CMD ["/usr/src/app/main"]
