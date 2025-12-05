package com.vira.dply.websocket;

import java.io.IOException;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicBoolean;

import org.springframework.stereotype.Component;
import org.springframework.web.socket.CloseStatus;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;
import org.springframework.web.socket.handler.TextWebSocketHandler;

import com.vira.dply.enums.WebsocketLogType;

@Component
public class LogsWebSocketHandler extends TextWebSocketHandler {

    private final Map<String, WebSocketSession> sessions = new ConcurrentHashMap<>();
    private final Map<String, AtomicBoolean> cancelFlags = new ConcurrentHashMap<>();

    @Override
    public void afterConnectionEstablished(WebSocketSession session) {
        String deployId = session.getUri().getQuery(); // ?deployId=...
        sessions.put(deployId, session);
        cancelFlags.put(deployId, new AtomicBoolean(false));
    }

    @Override
    public void afterConnectionClosed(WebSocketSession session, CloseStatus status) {
        String deployId = session.getUri().getQuery();
        sessions.remove(deployId);
        cancelFlags.computeIfPresent(deployId, (k, v) -> {
            v.set(true);
            return null;
        });
    }

    public void sendLog(String deployId, String logLine, WebsocketLogType type) {
        // type = "DEPLOY" o "APP"
        WebSocketSession session = sessions.get(deployId);
        if (session != null && session.isOpen()) {
            try {
                session.sendMessage(new TextMessage(type + "|" + logLine));
            } catch (IOException e) {
                e.printStackTrace();
            }
        }
    }

    public boolean isCancelled(String deployId) {
        return cancelFlags.getOrDefault(deployId, new AtomicBoolean(false)).get();
    }
}
