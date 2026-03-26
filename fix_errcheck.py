import os
import re

def process_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    if filepath.endswith('_test.go'):
        if 'github.com/stretchr/testify/require' not in content:
            content = content.replace('import (', 'import (\n\t"github.com/stretchr/testify/require"', 1)
        
        # Replace req, _ := http.NewRequest(...)
        content = re.sub(
            r'req, _ := http\.NewRequest\((.*?)\)',
            r'req, err := http.NewRequest(\1)\n\trequire.NoError(t, err)',
            content
        )
        
        # Replace token, _ := utils.GenerateToken(...)
        content = re.sub(
            r'token, _ := utils\.GenerateToken(.*?)\n',
            r'token, err := utils.GenerateToken\1\n\trequire.NoError(t, err)\n',
            content
        )
        
        # Replace hashedPwd, _ := utils.HashPassword(...)
        content = re.sub(
            r'hashedPwd, _ := utils\.HashPassword(.*?)\n',
            r'hashedPwd, err := utils.HashPassword\1\n\trequire.NoError(t, err)\n',
            content
        )
        
        # Replace body, _ := json.Marshal(...)
        content = re.sub(
            r'body, _ := json\.Marshal\((.*?)\)',
            r'body, err := json.Marshal(\1)\n	require.NoError(t, err)',
            content
        )
        
        # Replace testLogger, _ := zap.NewDevelopment()
        content = re.sub(
            r'testLogger, _ := zap\.NewDevelopment\(\)',
            r'testLogger, err := zap.NewDevelopment()\n\trequire.NoError(t, err)',
            content
        )
        
        # Replace json.Unmarshal(...) isolated calls
        content = re.sub(
            r'\tjson\.Unmarshal\((.*?)\)\n',
            r'\terr = json.Unmarshal(\1)\n\trequire.NoError(t, err)\n',
            content
        )

        content = re.sub(
            r'logger, _ := zap\.NewDevelopment\(\)',
            r'logger, err := zap.NewDevelopment()\n\trequire.NoError(t, err)',
            content
        )

        content = re.sub(
            r'db, _ := gorm\.Open\((.*?)\)',
            r'db, err := gorm.Open(\1)\n\trequire.NoError(t, err)',
            content
        )
        
        content = re.sub(
            r'\tdb\.AutoMigrate\((.*?)\)\n',
            r'\terr = db.AutoMigrate(\1)\n\trequire.NoError(t, err)\n',
            content
        )


    elif filepath.endswith('export.go'):
        # Replace writer.Write([]string{...})
        content = re.sub(
            r'\twriter\.Write\((.*?)\)\n',
            r'\tif err := writer.Write(\1); err != nil {\n\t\tlogger.Errorw("Failed to write CSV row", "error", err.Error())\n\t\treturn\n\t}\n',
            content
        )

    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

base_dir = r"c:\Users\Administrator\Documents\malik\TA\absensholat-api"
files_to_process = [
    os.path.join(base_dir, "middleware", "auth_test.go"),
    os.path.join(base_dir, "handlers", "auth_test.go"),
    os.path.join(base_dir, "handlers", "export.go")
]

for f in files_to_process:
    if os.path.exists(f):
        process_file(f)
        print(f"Processed {f}")
    else:
        print(f"File not found {f}")
