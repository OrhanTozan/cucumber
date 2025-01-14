package io.cucumber.createmeta;

import io.cucumber.messages.types.Ci;
import io.cucumber.messages.types.Git;
import io.cucumber.messages.types.Meta;
import org.junit.jupiter.api.Test;

import java.util.HashMap;

import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.matchesPattern;
import static org.junit.jupiter.api.Assertions.assertEquals;

class CreateMetaTest {
    @Test
    void it_provides_the_correct_tool_name() {
        Meta meta = CreateMeta.createMeta("cucumber-jvm", "3.2.1", new HashMap<>());
        assertEquals("cucumber-jvm", meta.getImplementation().getName());
    }

    @Test
    void it_provides_the_correct_tool_version() {
        Meta meta = CreateMeta.createMeta("cucumber-jvm", "3.2.1", new HashMap<>());
        assertEquals("3.2.1", meta.getImplementation().getVersion());
    }

    @Test
    void it_provides_the_correct_protocol_version() {
        Meta meta = CreateMeta.createMeta("cucumber-jvm", "3.2.1", new HashMap<>());
        assertThat(meta.getProtocolVersion(), matchesPattern("\\d+\\.\\d+\\.\\d+(-SNAPSHOT)?"));
    }

    @Test
    void it_provides_the_correct_jvm_version_and_name() {
        Meta meta = CreateMeta.createMeta("cucumber-jvm", "3.2.1", new HashMap<>());
        assertThat(meta.getRuntime().getName(), matchesPattern("(OpenJDK).*"));
    }

    @Test
    void it_detects_github_actions() {
        HashMap<String, String> env = new HashMap<String, String>() {{
            put("GITHUB_SERVER_URL", "https://github.company.com");
            put("GITHUB_REPOSITORY", "cucumber/cucumber-ruby");
            put("GITHUB_RUN_ID", "140170388");
            put("GITHUB_SHA", "the-revision");
            put("GITHUB_REF", "refs/tags/the-tag");
        }};
        Meta meta = CreateMeta.createMeta("cucumber-jvm", "3.2.1", env);

        assertEquals(new Ci(
                        "GitHub Actions",
                        "https://github.company.com/cucumber/cucumber-ruby/actions/runs/140170388",
                        new Git(
                                "https://github.company.com/cucumber/cucumber-ruby.git",
                                "the-revision",
                                null,
                                "the-tag"
                        )
                ),
                meta.getCi());
    }

    @Test
    void can_handle_CIs_with_null_in_ciDict() {
        // choose gocd as example, which has null everywhere except for url and revision
        HashMap<String, String> env = new HashMap<String, String>() {{
            put("GO_SERVER_URL", "https://<cihost>/buildurl");
            put("GO_REVISION", "the-revision");
        }};
        Meta meta = CreateMeta.createMeta("cucumber-jvm", "3.2.1", env);
        assertEquals(new Ci(
                        "GoCD",
                        "https://<cihost>/buildurl/???",
                        new Git(
                                null,
                                "the-revision",
                                null,
                                null
                        )
                ),
                meta.getCi());
    }
}
