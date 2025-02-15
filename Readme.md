Dead simple git commit message generator.
Should i want to generate a commit message, i feel very lazy adding it to staging and clicking the sparkly icon in cursor. So i guess i made this to save me some time (irrational time).

##

The plan will be to extend this a bit:

- Generate commit messages for all staged files with recursive mode
- Accept the commit message and apply it to the files / Accept all commit messages and apply them to the files
- Commit the changes with the generated commit messages.
- If no staged files, stage all files and generate a commit messages for them.

## Usage

```bash
export OPENAI_API_KEY="your-openai-api-key"

# run with either file name or recursive mode
go run . <file-name>
go run . --r
```

Pheww, learnt a lot of GoLang with this haha

Recursive mode now works. Hooray
