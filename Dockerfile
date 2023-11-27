FROM golang:alpine

# Install npm
RUN apk add --update nodejs npm

# Set working directory
WORKDIR /usr/src/app

# Install dependencies
RUN npm install

# Generate css
RUN npm run css-generate

# Install template engine
RUN go install github.com/a-h/templ/cmd/templ@latest

# Generate html
RUN ${GOPATH}/bin/templ

# Copy files
COPY . .

# Expose port
EXPOSE 3000

# Run the app
CMD ["go", "run", "main.go"]
