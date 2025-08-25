# Python 🐍

## 💻 Development
Inside of this directory, you will see a `src/` sub-directory with a `main.py` file. This will be your entrypoint for developing your code.
Make sure to maintain the `main.py` file and the `handler` function inside that file, since Lambda will look for that function when running.

If you want to structure your code in multiple files, you can create them inside of the `src/` subdirectory.
All files should be in the top-level of the `src/` sub-directory for deployment to work. Do not create any nested sub-directories inside `src/`

## ⚙️ Managing dependencies
Dependencies are automatically detected through the `requirements.txt` file and deployed with the scripts present in the repository.
If you were to use any dependencies other than the built-in Python libraries and boto3, make sure to add it to the `requirements.txt` file.

For local development, create a new virtual environment by running the following commands
```bash
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

Then, set up this virtual environment as your Python interpreter.