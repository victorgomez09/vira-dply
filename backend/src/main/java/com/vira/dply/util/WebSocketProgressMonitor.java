package com.vira.dply.util;

import org.eclipse.jgit.lib.ProgressMonitor;

import com.vira.dply.enums.WebsocketLogType;
import com.vira.dply.websocket.LogsWebSocketHandler;

public class WebSocketProgressMonitor implements ProgressMonitor {

    private final String deployId;
    private final LogsWebSocketHandler logsWebSocketHandler;

    public WebSocketProgressMonitor(String deployId, LogsWebSocketHandler logsWebSocketHandler) {
        this.deployId = deployId;
        this.logsWebSocketHandler = logsWebSocketHandler;
    }

    @Override
    public void start(int totalTasks) {}

    @Override
    public void beginTask(String title, int totalWork) {
        logsWebSocketHandler.sendLog(deployId, "[GIT] " + title, WebsocketLogType.DEPLOY);
    }

    @Override
    public void update(int completed) {
        // Opcional: mostrar progreso numérico
    }

    @Override
    public void endTask() {}

    @Override
    public boolean isCancelled() {
        return false;
    }

    @Override
    public void showDuration(boolean enabled) {
        // TODO Auto-generated method stub
    }
}