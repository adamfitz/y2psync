const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{}).result;

    const go_os = switch (target.os.tag) {
        .linux => "linux",
        .macos => "darwin",
        .windows => "windows",
        else => @panic("unsupported OS"),
    };

    const go_arch = switch (target.cpu.arch) {
        .x86_64 => "amd64",
        .aarch64 => "arm64",
        else => @panic("unsupported arch"),
    };

    const is_native = (std.mem.eql(u8, go_os, "linux") and std.mem.eql(u8, go_arch, "amd64")) or
        (std.mem.eql(u8, go_os, "darwin") and std.mem.eql(u8, go_arch, "arm64"));

    const exe_name = if (std.mem.eql(u8, go_os, "windows")) "y2psync.exe" else "y2psync";
    const out_dir = b.fmt("build/{s}-{s}", .{ go_os, go_arch });

    const os_tag_str: []const u8 = if (std.mem.eql(u8, go_os, "darwin")) "macos" else @tagName(target.os.tag);

    const go_build = b.addSystemCommand(&.{
        "go", "build",
        "-ldflags=-s -w -extldflags=-static-libgcc",
        "-trimpath",
        "-o", b.fmt("{s}/{s}", .{ out_dir, exe_name }),
        "./cmd/y2psync/",
    });
    go_build.setEnvironmentVariable("GOOS", go_os);
    go_build.setEnvironmentVariable("GOARCH", go_arch);
    go_build.setEnvironmentVariable("CGO_ENABLED", "1");

    if (!is_native) {
        const zig_target_raw = b.fmt("{s}-{s}", .{
            @tagName(target.cpu.arch),
            os_tag_str,
        });
        go_build.setEnvironmentVariable("CC", b.fmt("zig cc -target {s}", .{zig_target_raw}));
    }
    go_build.setEnvironmentVariable("PATH", "/usr/bin:/usr/local/bin:/bin");

    const build_step = b.step("build", "Build y2psync for the target OS/arch");
    build_step.dependOn(&go_build.step);

    const default_step = b.step("native", "Build y2psync for the native OS/arch");
    default_step.dependOn(&go_build.step);

    const clean_cmd = b.addSystemCommand(&.{ "rm", "-rf", "build/" });
    const clean_step = b.step("clean", "Remove build artifacts");
    clean_step.dependOn(&clean_cmd.step);
}
