package com.vira.dply.util;

import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;

public class BuildResultStore {

    private static final ConcurrentHashMap<String, CompletableFuture<BuildResult>> store = new ConcurrentHashMap<>();

    public static void add(String buildId, CompletableFuture<BuildResult> future) {
        store.put(buildId, future);
    }

    public static CompletableFuture<BuildResult> get(String buildId) {
        return store.get(buildId);
    }

    public static void remove(String buildId) {
        store.remove(buildId);
    }
}
