# MyWeb

MyWeb is a simple web application for managing blog posts, media files, and animations. 
It is built with Go and uses a bbolt + storm database for dynamic content.

## Features

- Blogs - A page for looking at blog descriptions to navigate to the full blog post.
- Media - Upload media files and automatically convert them to webp if possible.
- Animations - A page for displaying pictures and videos with short descriptions.
- Manage - A page for managing blog posts, media files, and animations.

## Recommendations

- Use a reverse proxy such as [Caddy](https://caddyserver.com/) to serve the application.
- Only listen on 127.0.0.1 or a local network address.
- Use a firewall to block all incoming traffic except for the reverse proxy.

## Installation

1. Set up configuration:
   - Edit the webhome/myweb.conf file.
   - Set the URL, maintainer name, and email.
   - Set the listen ip or port if needed.
   ```sh
   nano webhome/myweb.conf
   ```

2. Make a bcrypt password for the admin user:
   - Run the following command to create a password hash.
   - Replace "username" and "password" with your desired admin username and password.
    ```sh
   name="username" \
   pass="password" \
   caddy hash-password --plaintext="$name:$pass" > webhome/userpass
    ```
   - You can also use another implementation of bcrypt to create the password hash.
   - Format the file as `hash` where hash is the hash of the string "username:password".
   
3. Copy the webhome directory to your http user directory:
    ```sh
   # Copy files to http user directory. Use the actual user name for httpuser on your system.
   sudo cp -r ./webhome ~httpuser/
   # Change the owner of directory only, so the server can create new files.
   sudo chown httpuser ~httpuser/webhome
    ```

4. Build the application:
    ```sh
    go build -ldflags="-w -s" -trimpath -o build/
   ```

5. Install the application:
   ```sh
    sudo install ./build/myweb /usr/local/bin/myweb
   ```
   
## Systemd Service
   
1. Edit the service file:
   - Edit the `myweb.service` file.
   - Change the User to the user that your web server runs as.
   - Change the WorkingDirectory to the directory where you copied the webhome directory.
   - Change the ExecStart to the path where you installed the application if different.

2. Install the systemd service:
   ```sh 
   sudo install ./myweb.service /etc/systemd/system/myweb.service
   sudo systemctl daemon-reload
    ```

3. Start the service:
   ```sh
   sudo systemctl start myweb
   sudo systemctl enable myweb
   sudo systemctl status myweb
   ```
   
## Run from Command Line

```sh
  myweb /path/to/myweb.conf
```

## Manage Content

1. Log in to the manage page:
   - Go to the manage page at `/manage/`.
   - Log in with the username and password you created earlier.
   - Select management options from the menu.

## License

This project is licensed under the MIT License.