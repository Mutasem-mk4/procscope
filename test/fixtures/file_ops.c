/*
 * file_ops.c — Test fixture: performs various file operations.
 *
 * Used by procscope integration tests to verify:
 *   - file open (read and write modes)
 *   - file create
 *   - file rename
 *   - file delete
 *   - chmod
 *
 * Build: cc -o file_ops file_ops.c
 */

#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/stat.h>
#include <string.h>

int main(void) {
    const char *tmpdir = getenv("TMPDIR");
    if (!tmpdir) tmpdir = "/tmp";

    char path1[256], path2[256];
    snprintf(path1, sizeof(path1), "%s/procscope_test_file1.txt", tmpdir);
    snprintf(path2, sizeof(path2), "%s/procscope_test_file2.txt", tmpdir);

    /* Create and write a file */
    int fd = open(path1, O_WRONLY | O_CREAT | O_TRUNC, 0644);
    if (fd < 0) {
        perror("open(create)");
        return 1;
    }
    const char *data = "test data for procscope\n";
    write(fd, data, strlen(data));
    close(fd);
    printf("created %s\n", path1);

    /* Open for reading */
    fd = open(path1, O_RDONLY);
    if (fd < 0) {
        perror("open(read)");
        return 1;
    }
    char buf[256];
    read(fd, buf, sizeof(buf));
    close(fd);
    printf("read %s\n", path1);

    /* Chmod */
    if (chmod(path1, 0600) < 0) {
        perror("chmod");
    }
    printf("chmod 0600 %s\n", path1);

    /* Rename */
    if (rename(path1, path2) < 0) {
        perror("rename");
        return 1;
    }
    printf("renamed %s -> %s\n", path1, path2);

    /* Delete */
    if (unlink(path2) < 0) {
        perror("unlink");
        return 1;
    }
    printf("deleted %s\n", path2);

    return 0;
}
