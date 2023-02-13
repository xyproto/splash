#include <ctype.h>
#include <err.h>
#include <errno.h>
#include <getopt.h>
#include <inttypes.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/syscall.h>
#include <unistd.h>

static const char* version_string = "tinyionice 1.0.4";

enum {
    IOPRIO_CLASS_NONE,
    IOPRIO_CLASS_RT,
    IOPRIO_CLASS_BE,
    IOPRIO_CLASS_IDLE,
};
enum {
    IOPRIO_WHO_PROCESS = 1,
    IOPRIO_WHO_PGRP,
    IOPRIO_WHO_USER,
};

static int IOPRIO_CLASS_SHIFT = 13;
static int EX_EXEC_FAILED = 126; // Program located, but not usable
static int EX_EXEC_ENOENT = 127; // Could not find program to exec
static const char* to_prio[] = {
    [IOPRIO_CLASS_NONE] = "none",
    [IOPRIO_CLASS_RT] = "realtime",
    [IOPRIO_CLASS_BE] = "best-effort",
    [IOPRIO_CLASS_IDLE] = "idle"
};

static inline int flush_standard_stream(FILE* stream)
{
    errno = 0;
    if (ferror(stream) != 0 || fflush(stream) != 0) {
        return (errno == EBADF) ? 0 : EOF;
    }
    /*
     * Calling fflush is not sufficient on some filesystems like e.g. NFS,
     * which may defer the actual flush until close. Calling fsync would help
     * solve this, but would probably result in a performance hit. Thus, we
     * work around this issue by calling close on a dup'd file descriptor from
     * the stream.
     */
    int fd = fileno(stream);
    if (fd < 0 || (fd = dup(fd)) < 0 || close(fd) != 0) {
        return (errno == EBADF) ? 0 : EOF;
    }
    return 0;
}

/* Meant to be used atexit(close_stdout); */
static inline void close_stdout(void)
{
    if (flush_standard_stream(stdout) != 0 && !(errno == EPIPE)) {
        if (errno) {
            warn("write error");
        } else {
            warnx("write error");
        }
    } else if (flush_standard_stream(stderr) == 0) {
        return;
    }
    _exit(EXIT_FAILURE);
}

static inline int64_t strtos64_or_err(const char* str, const char* errmesg)
{
    char* end = NULL;
    int64_t num;
    errno = 0;
    if (str == NULL || *str == '\0') {
        if (errno == ERANGE) {
            err(EXIT_FAILURE, "%s: '%s'", errmesg, str);
        }
        errx(EXIT_FAILURE, "%s: '%s'", errmesg, str);
    }
    num = strtoimax(str, &end, 10);
    if (errno || str == end || (end && *end)) {
        if (errno == ERANGE) {
            err(EXIT_FAILURE, "%s: '%s'", errmesg, str);
        }
        errx(EXIT_FAILURE, "%s: '%s'", errmesg, str);
    }
    return num;
}

static inline int32_t strtos32_or_err(const char* str, const char* errmesg)
{
    int64_t num = strtos64_or_err(str, errmesg);
    if (num < INT32_MIN || num > INT32_MAX) {
        errno = ERANGE;
        err(EXIT_FAILURE, "%s: '%s'", errmesg, str);
    }
    return (int32_t)num;
}

static inline unsigned long IOPRIO_PRIO_MASK()
{
    return (1UL << IOPRIO_CLASS_SHIFT) - 1;
}

static inline unsigned long IOPRIO_PRIO_CLASS(unsigned long mask)
{
    return mask >> IOPRIO_CLASS_SHIFT;
}

static inline unsigned long IOPRIO_PRIO_DATA(unsigned long mask)
{
    return mask & IOPRIO_PRIO_MASK();
}

static inline unsigned long IOPRIO_PRIO_VALUE(unsigned long class, unsigned long data)
{
    return ((class << IOPRIO_CLASS_SHIFT) | data);
}

static int parse_ioclass(const char* str)
{
    for (int i = 0; i < 4; i++) {
        if (!strcasecmp(str, to_prio[i])) {
            return i;
        }
    }
    return -1;
}

static void ioprio_print(const int pid, const int who)
{
    const int ioprio = syscall(SYS_ioprio_get, who, pid);
    if (ioprio == -1) {
        err(EXIT_FAILURE, "ioprio_get failed");
    }
    const int ioclass = IOPRIO_PRIO_CLASS(ioprio);
    const char* name = "unknown";
    if (ioclass >= 0 && (size_t)ioclass < 4) {
        name = to_prio[ioclass];
    }
    if (ioclass != IOPRIO_CLASS_IDLE) {
        printf("%s: prio %lu\n", name, IOPRIO_PRIO_DATA(ioprio));
    } else {
        printf("%s\n", name);
    }
}

static void ioprio_setid(int which, int ioclass, int data, int who, bool tolerant)
{
    const int rc = syscall(SYS_ioprio_set, who, which, IOPRIO_PRIO_VALUE(ioclass, data));
    if (rc == -1 && !tolerant) {
        err(EXIT_FAILURE, "ioprio_set failed");
    }
}

static void __attribute__((__noreturn__)) usage(void)
{
    fputs("\nUsage:\n"
      " tinyionice [options] -p <pid>...\n tinyionice [options] -P <pgid>...\n"
      " tinyionice [options] -u <uid>...\n tinyionice [options] <command>\n\n"
      "Show or change the I/O-scheduling class and priority of a process.\n\n"
      "Options:\n"
      " -c, --class <class>    name or number of scheduling class,\n"
      "                          0: none, 1: realtime, 2: best-effort, 3: idle\n"
      " -n, --classdata <num>  priority (0..7) in the specified scheduling class,\n"
      "                          only for the realtime and best-effort classes\n"
      " -p, --pid <pid>...     act on these already running processes\n"
      " -P, --pgid <pgrp>...   act on already running processes in these groups\n"
      " -t, --ignore           ignore failures\n"
      " -u, --uid <uid>...     act on already running processes owned by these users\n\n"
      " -h, --help             display this help\n"
      " -V, --version          display version\n", stdout);
    exit(EXIT_SUCCESS);
}

int main(int argc, char** argv)
{
    bool tolerant = false;
    const char* invalid_msg = NULL;
    int data = 4, set = 0, c = 0, which = 0, who = 0, ioclass = IOPRIO_CLASS_BE;
    static const struct option longopts[] = {
        { "classdata", required_argument, NULL, 'n' },
        { "class", required_argument, NULL, 'c' },
        { "help", no_argument, NULL, 'h' },
        { "ignore", no_argument, NULL, 't' },
        { "pid", required_argument, NULL, 'p' },
        { "pgid", required_argument, NULL, 'P' },
        { "uid", required_argument, NULL, 'u' },
        { "version", no_argument, NULL, 'V' },
        { NULL, 0, NULL, 0 }
    };
    atexit(close_stdout);
    while ((c = getopt_long(argc, argv, "+n:c:p:P:u:tVh", longopts, NULL)) != EOF)
        switch (c) {
        case 'n':
            data = strtos32_or_err(optarg, "invalid class data argument");
            set |= 1;
            break;
        case 'c':
            if (isdigit(*optarg)) {
                ioclass = strtos32_or_err(optarg, "invalid class argument");
            } else {
                ioclass = parse_ioclass(optarg);
                if (ioclass < 0) {
                    errx(EXIT_FAILURE, "unknown scheduling class: '%s'", optarg);
                }
            }
            set |= 2;
            break;
        case 'p':
            if (who) {
                errx(EXIT_FAILURE, "can handle only one of pid, pgid or uid at once");
            }
            invalid_msg = "invalid PID argument";
            which = strtos32_or_err(optarg, invalid_msg);
            who = IOPRIO_WHO_PROCESS;
            break;
        case 'P':
            if (who) {
                errx(EXIT_FAILURE, "can handle only one of pid, pgid or uid at once");
            }
            invalid_msg = "invalid PGID argument";
            which = strtos32_or_err(optarg, invalid_msg);
            who = IOPRIO_WHO_PGRP;
            break;
        case 'u':
            if (who) {
                errx(EXIT_FAILURE, "can handle only one of pid, pgid or uid at once");
            }
            invalid_msg = "invalid UID argument";
            which = strtos32_or_err(optarg, invalid_msg);
            who = IOPRIO_WHO_USER;
            break;
        case 't':
            tolerant = 1;
            break;
        case 'V':
            printf("%s\n", version_string);
            exit(EXIT_SUCCESS);
        case 'h':
            usage();
        default:
            fprintf(stderr, "Try 'tinyionice --help' for more information.\n");
            exit(EXIT_FAILURE);
        }

    switch (ioclass) {
    case IOPRIO_CLASS_NONE:
        if ((set & 1) && !tolerant) {
            warnx("ignoring given class data for none class");
        }
        data = 0;
        break;
    case IOPRIO_CLASS_RT:
    case IOPRIO_CLASS_BE:
        break;
    case IOPRIO_CLASS_IDLE:
        if ((set & 1) && !tolerant) {
            warnx("ignoring given class data for idle class");
        }
        data = 7;
        break;
    default:
        if (!tolerant) {
            warnx("unknown prio class %d", ioclass);
        }
        break;
    }
    if (!set && !which && optind == argc) {
        // tinyionice without options, print the current ioprio
        ioprio_print(0, IOPRIO_WHO_PROCESS);
    } else if (!set && who) {
        // tinyionice -p|-P|-u ID [ID ...]
        ioprio_print(which, who);
        for (; argv[optind]; ++optind) {
            which = strtos32_or_err(argv[optind], invalid_msg);
            ioprio_print(which, who);
        }
    } else if (set && who) {
        // tinyionice -c CLASS -p|-P|-u ID [ID ...]
        ioprio_setid(which, ioclass, data, who, tolerant);
        for (; argv[optind]; ++optind) {
            which = strtos32_or_err(argv[optind], invalid_msg);
            ioprio_setid(which, ioclass, data, who, tolerant);
        }
    } else if (argv[optind]) {
        // tinyionce [-c CLASS] COMMAND
        ioprio_setid(0, ioclass, data, IOPRIO_WHO_PROCESS, tolerant);
        execvp(argv[optind], &argv[optind]);
        err(errno == ENOENT ? EX_EXEC_ENOENT : EX_EXEC_FAILED, "failed to execute %s", argv[optind]);
    } else {
        warnx("bad usage");
        fprintf(stderr, "Try 'tinyionice --help' for more information.\n");
        exit(EXIT_FAILURE);
    }
    return EXIT_SUCCESS;
}
