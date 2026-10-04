/*
 * fork_exec.c — Test fixture: forks a child that execs /bin/echo.
 *
 * Used by procscope integration tests to verify:
 *   - fork event detection
 *   - exec event detection
 *   - exit event detection
 *   - parent/child relationship tracking
 *
 * Build: cc -o fork_exec fork_exec.c
 */

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/wait.h>

int main(void) {
    printf("parent pid=%d\n", getpid());

    pid_t child = fork();
    if (child < 0) {
        perror("fork");
        return 1;
    }

    if (child == 0) {
        /* Child process */
        printf("child pid=%d\n", getpid());
        execl("/bin/echo", "echo", "hello from child", NULL);
        perror("execl");
        _exit(127);
    }

    /* Parent waits for child */
    int status;
    waitpid(child, &status, 0);
    printf("child exited with status %d\n", WEXITSTATUS(status));

    return 0;
}
