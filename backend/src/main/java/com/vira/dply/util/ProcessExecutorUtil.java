package com.vira.dply.util;

import java.nio.file.Path;
import java.util.List;

import org.springframework.stereotype.Component;

@Component
public class ProcessExecutorUtil {

    public void run(List<String> command) {
        try {
            Process process = new ProcessBuilder(command)
                    .inheritIO()
                    .start();

            if (process.waitFor() != 0) {
                throw new RuntimeException("Command failed: " + command);
            }
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public void run(List<String> command, Path workDir) {
        try {
            Process process = new ProcessBuilder(command)
                    .directory(workDir.toFile())
                    .inheritIO()
                    .start();

            if (process.waitFor() != 0) {
                throw new RuntimeException("Command failed: " + command);
            }
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }
}