const std = @import("std");

fn defaultArch(os: []const u8) []const u8 {
    if (std.mem.eql(u8, os, "mac") or std.mem.eql(u8, os, "darwin")) return "arm64";
    return "amd64";
}

fn toZigArch(arch: []const u8) std.Target.Cpu.Arch {
    if (std.mem.eql(u8, arch, "amd64") or std.mem.eql(u8, arch, "x86_64")) return .x86_64;
    if (std.mem.eql(u8, arch, "arm64") or std.mem.eql(u8, arch, "aarch64")) return .aarch64;
    @panic("unsupported arch, use: amd64, x86_64, arm64, aarch64");
}

fn toZigOs(os: []const u8) std.Target.Os.Tag {
    if (std.mem.eql(u8, os, "mac") or std.mem.eql(u8, os, "darwin") or std.mem.eql(u8, os, "macos")) return .macos;
    if (std.mem.eql(u8, os, "linux")) return .linux;
    if (std.mem.eql(u8, os, "windows")) return .windows;
    @panic("unsupported OS, use: linux, windows, mac, darwin");
}

fn toGoOs(os: std.Target.Os.Tag) []const u8 {
    return switch (os) {
        .linux => "linux",
        .macos => "darwin",
        .windows => "windows",
        else => @panic("unsupported OS"),
    };
}

fn toGoArch(arch: std.Target.Cpu.Arch) []const u8 {
    return switch (arch) {
        .x86_64 => "amd64",
        .aarch64 => "arm64",
        else => @panic("unsupported arch"),
    };
}

pub fn build(b: *std.Build) void {
    const os_opt = b.option([]const u8, "os", "Target OS: linux, windows, mac (or darwin)");
    const arch_opt = b.option([]const u8, "arch", "Target arch: amd64, x86_64, arm64, aarch64");

    const resolved = if (os_opt) |os| blk: {
        const arch_str = arch_opt orelse defaultArch(os);
        const query = std.Target.Query{
            .cpu_arch = toZigArch(arch_str),
            .os_tag = toZigOs(os),
        };
        break :blk b.resolveTargetQuery(query);
    } else if (arch_opt != null) {
        @panic("`-Darch` requires `-Dos` to also be set");
    } else b.standardTargetOptions(.{});

    const target = resolved.result;
    const go_os = toGoOs(target.os.tag);
    const go_arch = toGoArch(target.cpu.arch);

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

    const is_native = (std.mem.eql(u8, go_os, "linux") and std.mem.eql(u8, go_arch, "amd64")) or
        (std.mem.eql(u8, go_os, "darwin") and std.mem.eql(u8, go_arch, "arm64"));

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
