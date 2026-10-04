/*
 * net_connect.c — Test fixture: connects to localhost TCP.
 *
 * Used by procscope integration tests to verify:
 *   - bind event
 *   - listen event
 *   - connect event
 *   - accept event
 *
 * Creates a local TCP server, connects to it, then closes.
 *
 * Build: cc -o net_connect net_connect.c
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

int main(void) {
    /* Create server socket */
    int server_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (server_fd < 0) {
        perror("socket(server)");
        return 1;
    }

    int opt = 1;
    setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));

    struct sockaddr_in addr;
    memset(&addr, 0, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = inet_addr("127.0.0.1");
    addr.sin_port = 0; /* kernel assigns port */

    if (bind(server_fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("bind");
        close(server_fd);
        return 1;
    }
    printf("bind ok\n");

    /* Get assigned port */
    socklen_t addrlen = sizeof(addr);
    getsockname(server_fd, (struct sockaddr *)&addr, &addrlen);
    int port = ntohs(addr.sin_port);
    printf("listening on 127.0.0.1:%d\n", port);

    if (listen(server_fd, 5) < 0) {
        perror("listen");
        close(server_fd);
        return 1;
    }
    printf("listen ok\n");

    /* Fork: child connects, parent accepts */
    pid_t child = fork();
    if (child < 0) {
        perror("fork");
        close(server_fd);
        return 1;
    }

    if (child == 0) {
        /* Child: connect to server */
        close(server_fd);
        usleep(50000); /* 50ms delay for server readiness */

        int client_fd = socket(AF_INET, SOCK_STREAM, 0);
        if (client_fd < 0) {
            perror("socket(client)");
            _exit(1);
        }

        struct sockaddr_in srv_addr;
        memset(&srv_addr, 0, sizeof(srv_addr));
        srv_addr.sin_family = AF_INET;
        srv_addr.sin_addr.s_addr = inet_addr("127.0.0.1");
        srv_addr.sin_port = htons(port);

        if (connect(client_fd, (struct sockaddr *)&srv_addr, sizeof(srv_addr)) < 0) {
            perror("connect");
            close(client_fd);
            _exit(1);
        }
        printf("connect ok to 127.0.0.1:%d\n", port);

        close(client_fd);
        _exit(0);
    }

    /* Parent: accept connection */
    struct sockaddr_in client_addr;
    socklen_t client_len = sizeof(client_addr);
    int accepted_fd = accept(server_fd, (struct sockaddr *)&client_addr, &client_len);
    if (accepted_fd < 0) {
        perror("accept");
    } else {
        printf("accepted connection from %s:%d\n",
               inet_ntoa(client_addr.sin_addr), ntohs(client_addr.sin_port));
        close(accepted_fd);
    }

    /* Wait for child */
    int status;
    waitpid(child, &status, 0);

    close(server_fd);
    return 0;
}
