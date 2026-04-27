import os

def check_and_convert(file_path):
    with open(file_path, 'rb') as f:
        content = f.read()
    
    if b'\r\n' in content:
        print(f"Converting {file_path} to LF...")
        new_content = content.replace(b'\r\n', b'\n')
        with open(file_path, 'wb') as f:
            f.write(new_content)
        return True
    return False

files_to_check = [
    'arch/PKGBUILD',
    'arch/.SRCINFO',
    'Makefile',
]

# Add scripts and debian files
for root, dirs, files in os.walk('.'):
    if '.git' in dirs:
        dirs.remove('.git')
    for file in files:
        if root.startswith('./scripts') or root.startswith('./debian') or file.endswith('.sh'):
            files_to_check.append(os.path.join(root, file))

converted_count = 0
for f in set(files_to_check):
    if os.path.isfile(f):
        if check_and_convert(f):
            converted_count += 1

print(f"Total files converted: {converted_count}")
