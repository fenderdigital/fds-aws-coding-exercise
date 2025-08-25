# Node.js 🐢🚀

## 💻 Development
Inside of this directory, you will see a `src/` sub-directory with an `index.js` file. This will be your entrypoint for developing your code.
Make sure to maintain the `index.js` file and the `exports.handler` function inside that file, since Lambda will look for that function when running.

If you want to structure your code in multiple files, you can create them inside of the `src/` subdirectory.
All files should be in the top-level of the `src/` sub-directory for deployment to work. Do not create any nested sub-directories inside `src/`

## ⚙️ Managing dependencies
Dependencies are automatically detected through `npm` and deployed with the scripts present in the repository.
You can add external dependencies with `npm i <dependency name>` and they will get added to the `node_modules/` subdirectory.

The deployment script automatically installs all dependencies in the `package.json` file.