# Go 🦫

## 💻 Development
Inside of this directory, you will see a `src/` sub-directory with a `main.go` file. This will be your entrypoint for developing your code.
Make sure to maintain the `main.go` file and both the `handler` and `main` functions inside that file, since Lambda will look for those functions when running.

If you want to structure your code in multiple files, you can create them inside of the `src/` subdirectory.

## ⚙️ Managing dependencies
To manage dependencies use `go vendor` to create a `vendor` directory with the dependency code.

Once you add an external dependency to your code, run the following commands to set up the vendor.

```bash
go mod tidy
go vendor
```

The deployment script will automatically add all dependencies to the `vendor` directory.