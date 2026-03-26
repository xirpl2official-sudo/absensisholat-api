import os
import re

def process_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    if filepath.endswith('_test.go'):
        if 'github.com/stretchr/testify/require' not in content:
            content = content.replace('import (', 'import (\n\t"github.com/stretchr/testify/require"\n', 1)
        
        content = re.sub(
            r'req, _ := http\.NewRequest\((.*?)\)',
            r'req, err := http.NewRequest(\1)\n\trequire.NoError(t, err)',
            content
        )
        
        content = re.sub(
            r'token, _ := utils\.GenerateToken(.*?)\n',
            r'token, err := utils.GenerateToken\1\n\trequire.NoError(t, err)\n',
            content
        )

        content = re.sub(
            r'token, _ := utils\.GenerateTokenWithNIP(.*?)\n',
            r'token, err := utils.GenerateTokenWithNIP\1\n\trequire.NoError(t, err)\n',
            content
        )
        
        content = re.sub(
            r'hashedPwd, _ := utils\.HashPassword(.*?)\n',
            r'hashedPwd, err := utils.HashPassword\1\n\trequire.NoError(t, err)\n',
            content
        )
        
        content = re.sub(
            r'body, _ := json\.Marshal\((.*?)\)',
            r'body, err := json.Marshal(\1)\n\trequire.NoError(t, err)',
            content
        )
        
        content = re.sub(
            r'testLogger, _ := zap\.NewDevelopment\(\)',
            r'testLogger, err := zap.NewDevelopment()\n\trequire.NoError(t, err)',
            content
        )

        # Skip json.Unmarshal for now as `err` scope could clash.
        # Instead, replace `json.Unmarshal(w.Body.Bytes(), &response)` with `_ = json.Unmarshal(...)`
        content = re.sub(
            r'\tjson\.Unmarshal\((.*?)\)\n',
            r'\t_ = json.Unmarshal(\1)\n',
            content
        )

        content = re.sub(
            r'logger, _ := zap\.NewDevelopment\(\)',
            r'logger, err := zap.NewDevelopment()\n\trequire.NoError(t, err)',
            content
        )
        
        content = re.sub(
            r'\tdb\.AutoMigrate\((.*?)\)\n',
            r'\terr = db.AutoMigrate(\1)\n\trequire.NoError(t, err)\n',
            content
        )

    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

base_dir = r"c:\Users\Administrator\Documents\malik\TA\absensholat-api"
files_to_process = [
    os.path.join(base_dir, "middleware", "auth_test.go"),
    os.path.join(base_dir, "handlers", "auth_test.go")
]

for f in files_to_process:
    if os.path.exists(f):
        process_file(f)
        print(f"Processed {f}")
