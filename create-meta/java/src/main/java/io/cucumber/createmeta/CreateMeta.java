package io.cucumber.createmeta;

import com.eclipsesource.json.Json;
import com.eclipsesource.json.JsonObject;
import com.eclipsesource.json.JsonValue;
import io.cucumber.messages.ProtocolVersion;
import io.cucumber.messages.types.Ci;
import io.cucumber.messages.types.Git;
import io.cucumber.messages.types.Meta;
import io.cucumber.messages.types.Product;

import java.io.IOException;
import java.io.InputStreamReader;
import java.io.Reader;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import static java.nio.charset.StandardCharsets.UTF_8;

public class CreateMeta {
    private static final JsonObject CI_DICT;
    private static final String JSON_PATH = "/io/cucumber/createmeta/ciDict.json";

    static {
        try (Reader reader = new InputStreamReader(CreateMeta.class.getResourceAsStream(JSON_PATH), UTF_8)) {
            CI_DICT = Json.parse(reader).asObject();
        } catch (IOException e) {
            throw new RuntimeException("Unable to parse " + JSON_PATH, e);
        }
    }

    public static Meta createMeta(
            String implementationName,
            String implementationVersion,
            Map<String, String> env
    ) {
        return new Meta(
                ProtocolVersion.getVersion(),
                new Product(implementationName, implementationVersion),
                new Product(System.getProperty("java.vm.name"), System.getProperty("java.vm.version")),
                new Product(System.getProperty("os.name"), null),
                new Product(System.getProperty("os.arch"), null),
                detectCI(env)
        );
    }

    public static String removeUserInfoFromUrl(String value) {
        if (value == null) return null;
        try {
            URI uri = URI.create(value);
            return new URI(uri.getScheme(), null, uri.getHost(), uri.getPort(), uri.getPath(), uri.getQuery(), uri.getFragment()).toASCIIString();
        } catch (URISyntaxException | IllegalArgumentException e) {
            return value;
        }
    }

    static Ci detectCI(Map<String, String> env) {
        List<Ci> detected = new ArrayList<>();
        for (JsonObject.Member envEntry : CI_DICT) {
            Ci ci = createCi(envEntry.getName(), envEntry.getValue().asObject(), env);
            if (ci != null) {
                detected.add(ci);
            }
        }
        return detected.size() == 1 ? detected.get(0) : null;
    }

    private static Ci createCi(String name, JsonObject ciSystem, Map<String, String> env) {
        String url = evaluate(getString(ciSystem, "url"), env);
        if (url == null) return null;
        JsonObject git = ciSystem.get("git").asObject();
        String remote = removeUserInfoFromUrl(evaluate(getString(git, "remote"), env));
        String revision = evaluate(getString(git, "revision"), env);
        String branch = evaluate(getString(git, "branch"), env);
        String tag = evaluate(getString(git, "tag"), env);

        return new Ci(
                name,
                url,
                new Git(remote, revision, branch, tag)
        );
    }

    private static String evaluate(String template, Map<String, String> env) {
        if (template == null) return null;
        try {
            Pattern pattern = Pattern.compile("\\$\\{((refbranch|reftag)\\s+)?([^\\s}]+)(\\s+\\|\\s+([^}]+))?}");
            Matcher matcher = pattern.matcher(template);
            StringBuffer sb = new StringBuffer();
            while (matcher.find()) {
                String func = matcher.group(2);
                String variable = matcher.group(3);
                String defaultValue = matcher.group(5);
                String value = env.getOrDefault(variable, defaultValue);
                if (value == null) {
                    throw new RuntimeException(String.format("Undefined variable: %s", variable));
                }
                if (func != null) {
                    switch (func) {
                        case "refbranch":
                            value = group1(value, Pattern.compile("^refs/heads/(.*)"));
                            break;
                        case "reftag":
                            value = group1(value, Pattern.compile("^refs/tags/(.*)"));
                            break;
                    }
                }
                if (value == null) {
                    throw new RuntimeException(String.format("Undefined variable: %s", variable));
                }
                matcher.appendReplacement(sb, value);
            }
            matcher.appendTail(sb);
            return sb.toString();
        } catch (RuntimeException e) {
            return null;
        }
    }

    private static String group1(String value, Pattern pattern) {
        Matcher matcher = pattern.matcher(value);
        if (matcher.find()) {
            return matcher.group(1);
        }
        return matcher.find() ? matcher.group(1) : null;
    }

    private static String getString(JsonObject json, String name) {
        JsonValue val = json.get(name);
        return val.isNull() ? null : val.asString();
    }
}
