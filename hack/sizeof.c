/* SPDX-License-Identifier: BSD-2-Clause */

#include <sys/ioctl.h>
#include <sys/syscall.h>
#include <linux/userfaultfd.h>
#include <stdio.h>

#define PRINT_SIZE(s)	printf("%-20s %zu\n", #s, sizeof(struct s))
#define PRINT_VAL(s)	printf("%-20s %lx\n", #s, s)

int main(void)
{
	PRINT_VAL(UFFDIO_REGISTER);
	PRINT_VAL(UFFDIO_UNREGISTER);
	PRINT_VAL(UFFDIO_WAKE);
	PRINT_VAL(UFFDIO_COPY);
	PRINT_VAL(UFFDIO_ZEROPAGE);
	PRINT_VAL(UFFDIO_MOVE);
	PRINT_VAL(UFFDIO_WRITEPROTECT);
	PRINT_VAL(UFFDIO_CONTINUE);
	PRINT_VAL(UFFDIO_POISON);
	PRINT_SIZE(uffd_msg);
	PRINT_SIZE(uffdio_api);
	PRINT_SIZE(uffdio_range);
	PRINT_SIZE(uffdio_register);
	PRINT_SIZE(uffdio_copy);
	PRINT_SIZE(uffdio_zeropage);
	PRINT_SIZE(uffdio_writeprotect);
	PRINT_SIZE(uffdio_continue);
	PRINT_SIZE(uffdio_poison);
	PRINT_SIZE(uffdio_move);

	return 0;
}

