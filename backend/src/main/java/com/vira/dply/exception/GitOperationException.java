package com.vira.dply.exception;

public class GitOperationException extends RuntimeException {
    public GitOperationException(String message, Throwable cause) {
        super(message, cause);
    }
}