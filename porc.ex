defmodule Porc do
  defrecord Process, [:port, :in, :out, :err]

  @doc """
  Takes a shell invocation and produces a tuple `{ cmd, args }` suitable for
  use in `execute()` and `spawn()` functions.
  """
  def shplit(invocation) when is_binary(invocation) do
    [cmd, rest] = String.split invocation, " ", global: false
    { cmd, split(rest) }
  end

  defp split(args) when is_binary(args) do
    String.split args, " "
  end

  @doc """
  Executes the command synchronously. Takes the same options as `spawn()`
  except for one difference: `options[:in]` cannot be `:pid`; must be either a
  binary or `{ :file, <file> }`.
  """
  def execute(cmd, options) when is_binary(cmd) do
    execute(shplit(cmd), options)
  end

  def execute({ cmd, args }, options) when is_binary(cmd)
                                       and is_binary(args) do
    execute({ cmd, split(args) }, options)
  end

  def execute({ cmd, args }, options) when is_binary(cmd)
                                       and is_list(args)
                                       and is_list(options) do
    if options[:in] == :pid do
      raise RuntimeError, message: "Option [in: :pid] cannot be used with execute()"
    end

    { :ok, ccmd } = String.to_char_list(cmd)
    port_opts = port_options(options, args)
    port = Port.open { :spawn_executable, ccmd }, port_opts

    input  = process_input_opts(port, options[:in])
    output = process_output_opts(port, options[:out])
    error  = process_error_opts(port, options[:err])

    communicate(port, input, output, error)
  end

  defp communicate(port, input, output, _error) do
    if input do
      IO.puts "Passing input to port: #{input}"
      Port.command(port, input)
      Port.close(port)
    end
    collect_output(port, output, nil, nil, 0)
  end

  defp collect_output(port, output, out_data, err_data, status) do
    IO.puts "Collecting output"
    receive do
      { ^port, {:data, data} } ->
        out_data = process_port_output(output, data, out_data)
        IO.puts "Got data #{inspect out_data}"
        #collect_output(port, output, out_data, err_data, status)
        { 0, flatten(out_data), err_data }

      { ^port, {:exit_status, status} } ->
        { status, out_data, err_data }

      #{ ^port, :eof } ->
        #collect_output(port, output, out_data, err_data, true, did_see_exit, status)
    end
  end

  defp process_port_output(output_opt, in_data, nil) do
    process_port_output(output_opt, in_data, "")
  end

  defp process_port_output({ :buffer, _ }, in_data, out_data) do
    [out_data, in_data]
  end

  defp process_port_output({:file, file}, in_data, out_data) do
    :ok = IO.write file, in_data
    out_data
  end

  defp process_port_output(pid, in_data, out_data) when is_pid(pid) do
    pid <- in_data
    out_data
  end

  @doc """
  Spawn an external process and returns `Process` record ready for
  communication.
  """
  def spawn(cmd, options) when is_binary(cmd) do
    spawn(shplit(cmd), options)
  end

  def spawn({ cmd, args }, options) when is_binary(cmd)
                                     and is_binary(args) do
    spawn({ cmd, split(args) }, options)
  end

  def spawn({ cmd, args }, options) when is_binary(cmd)
                                     and is_list(args)
                                     and is_list(options) do
    { :ok, ccmd } = String.to_char_list(cmd)
    port_opts = port_options(options, args)
    port = Port.open { :spawn_executable, ccmd }, port_opts

    input  = process_input_opts(port, options[:in])
    output = process_output_opts(port, options[:out])
    error  = process_error_opts(port, options[:err])

    Process[port: port, in: input, out: output, err: error]
  end

  defp port_options(options, args) do
    port_opts = [{:args, args}, :stream, :use_stdio, :binary, :exit_status, :hide]
    if options[:err] == :out do
      port_opts = [:stderr_to_stdio | port_opts]
    end
    port_opts
  end

  defp process_input_opts(port, opt) do
    case opt do
      nil          -> nil
      :pid         -> :something # TODO: spawn(...)
      { :file, f } -> { :file, f }
      bin when is_binary(bin) ->
        # TODO: make it async
        Port.command(port, bin)
        nil
    end
  end

  defp process_output_opts(_port, opt) do
    case opt do
      nil          -> { :buffer, nil }
      :err         -> nil
      :buffer      -> { :buffer, nil }
      { :file, f } -> { :file, f }
      pid when is_pid(pid) ->
        pid
    end
  end

  defp process_error_opts(_port, opt) do
    case opt do
      nil          -> { :buffer, nil }
      :out         -> nil
      :buffer      -> { :buffer, nil }
      { :file, f } -> { :file, f }
      pid when is_pid(pid) ->
        pid
    end
  end


  defp flatten(list) do
    flatten(list, [])
  end

  defp flatten([], acc) do
    acc |> Enum.reverse |> Enum.join("")
  end

  defp flatten([ [] | t ], acc) do
    flatten(t, acc)
  end

  defp flatten([ [h|t] | tt ], acc) do
    flatten([ t | tt ], [h|acc])
  end

  defp flatten([ h | t ], acc) do
    flatten(t, [h|acc])
  end
end

#p = Porc.spawn("cat", in: "abc" | :pid | {:file, ...},
                     #out: :err | :buffer | pid | {:file, ...},
                     #err: :out | :buffer | pid | {:file, ...})

#p = Porc.execute("cat", in: "Hello world!", out: :buffer)

p = Port.open({:spawn_executable, '/usr/local/bin/go'}, [{:args, ["run", "main.go", "cat"]}, :binary, {:packet, 2}, :exit_status])
p = Port.open({:spawn_executable, '/bin/cat'}, [:binary, :stream, :exit_status])
