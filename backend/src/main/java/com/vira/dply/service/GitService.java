package com.vira.dply.service;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.KeyPair;
import java.util.UUID;
import java.util.function.Function;

import org.eclipse.jgit.api.CloneCommand;
import org.eclipse.jgit.api.Git;
import org.eclipse.jgit.transport.SshTransport;
import org.eclipse.jgit.transport.UsernamePasswordCredentialsProvider;
import org.eclipse.jgit.transport.sshd.KeyCache;
import org.eclipse.jgit.transport.sshd.SshdSessionFactory;
import org.eclipse.jgit.transport.sshd.SshdSessionFactoryBuilder;
import org.springframework.stereotype.Service;

import com.vira.dply.exception.GitOperationException;
import com.vira.dply.util.GitCredentials;
import com.vira.dply.util.InMemoryKeyLoader;
import com.vira.dply.util.WebSocketProgressMonitor;
import com.vira.dply.websocket.LogsWebSocketHandler;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@Service
@RequiredArgsConstructor
public class GitService {

    private static final Path BASE_DIR = Path.of("/tmp/paas/git");

    private final InMemoryKeyLoader inMemoryKeyLoader;
    private final LogsWebSocketHandler logsWebSocketHandler;

    public Path cloneRepo(
            String deployId,
            String repositoryUrl,
            String branch,
            GitCredentials credentials) {
        try {
            Files.createDirectories(BASE_DIR);
            Path targetDir = BASE_DIR.resolve(UUID.randomUUID().toString());

            CloneCommand clone = Git.cloneRepository()
                    .setURI(repositoryUrl)
                    .setBranch(branch)
                    .setDirectory(targetDir.toFile())
                    .setCloneAllBranches(false)
                    .setProgressMonitor(new WebSocketProgressMonitor(deployId, logsWebSocketHandler))
                    .setDepth(1);

            if (repositoryUrl.startsWith("http")) {
                configureHttps(clone, credentials);
            } else if (repositoryUrl.startsWith("git@") || repositoryUrl.startsWith("ssh://")) {
                configureSsh(clone, credentials);
            }

            log.info("Cloning {} [{}]", repositoryUrl, branch);
            clone.call();

            return targetDir;

        } catch (Exception e) {
            throw new GitOperationException("Failed to clone repository", e);
        }
    }

    /* ================= HTTPS ================= */
    private void configureHttps(CloneCommand clone, GitCredentials creds) {
        if (creds == null || creds.passwordOrToken() == null)
            return;

        clone.setCredentialsProvider(
                new UsernamePasswordCredentialsProvider(
                        creds.username() != null ? creds.username() : "git",
                        creds.passwordOrToken()));
    }

    /* ================= SSH (Apache MINA) ================= */
    private void configureSsh(CloneCommand clone, GitCredentials creds) {
        if (creds == null || creds.privateKey() == null) {
            throw new IllegalArgumentException("SSH private key is required");
        }

        // Proveedor de claves que lee desde memoria / string
        Function<File, Iterable<KeyPair>> keyProvider = sshDir -> {
            return inMemoryKeyLoader.load(
                    creds.privateKey(),
                    creds.passphrase());
        };

        SshdSessionFactory factory = new SshdSessionFactoryBuilder()
                .setPreferredAuthentications("publickey")
                .setDefaultKeysProvider(keyProvider)
                // opcional: .setHomeDirectory(...) / .setSshDirectory(...)
                .build((KeyCache) null);

        clone.setTransportConfigCallback(tr -> {
            ((SshTransport) tr).setSshSessionFactory(factory);
        });
    }

}
