package com.vira.dply.util;

import java.io.IOException;
import java.net.ServerSocket;

import org.springframework.stereotype.Component;

@Component
public final class PortUtil {

    public int findFreePort() {
        try (ServerSocket socket = new ServerSocket(0)) {
            socket.setReuseAddress(true);
            return socket.getLocalPort();
        } catch (IOException e) {
            throw new RuntimeException("Unable to find free port", e);
        }
    }
}